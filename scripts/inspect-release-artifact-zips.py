#!/usr/bin/env python3
import argparse
import binascii
import stat
import struct
import sys
import zipfile
import zlib


CHUNK_SIZE = 1024 * 1024
LOCAL_HEADER = struct.Struct("<IHHHHHIIIHH")
LOCAL_HEADER_SIGNATURE = 0x04034B50
DATA_DESCRIPTOR_SIGNATURE = 0x08074B50
UINT32_MAX = (1 << 32) - 1


def fail(message):
    raise ValueError(message)


def read_exact(source, size):
    data = source.read(size)
    if len(data) != size:
        fail("truncated ZIP structure")
    return data


def local_record(source, entry, central_offset):
    if entry.header_offset < 0 or entry.header_offset + LOCAL_HEADER.size > central_offset:
        fail(f"invalid local header offset: {entry.filename!r}")
    if entry.compress_size > UINT32_MAX or entry.file_size > UINT32_MAX:
        fail(f"unsupported ZIP64 member: {entry.filename!r}")

    source.seek(entry.header_offset)
    header = LOCAL_HEADER.unpack(read_exact(source, LOCAL_HEADER.size))
    signature, _, flags, method, _, _, crc, compressed_size, file_size, name_size, extra_size = header
    if signature != LOCAL_HEADER_SIGNATURE:
        fail(f"invalid local header: {entry.filename!r}")
    if flags != entry.flag_bits or method != entry.compress_type:
        fail(f"local header mismatch: {entry.filename!r}")
    encoding = "utf-8" if flags & 0x800 else "cp437"
    local_name = read_exact(source, name_size).decode(encoding)
    if local_name != entry.orig_filename:
        fail(f"local filename mismatch: {entry.filename!r}")
    read_exact(source, extra_size)
    data_start = source.tell()
    data_end = data_start + entry.compress_size
    if data_start < entry.header_offset or data_end < data_start or data_end > central_offset:
        fail(f"invalid compressed range: {entry.filename!r}")

    if flags & 0x08:
        if crc not in (0, entry.CRC) or compressed_size not in (0, entry.compress_size) or file_size not in (0, entry.file_size):
            fail(f"local descriptor fields mismatch: {entry.filename!r}")
        source.seek(data_end)
        first = read_exact(source, 4)
        if struct.unpack("<I", first)[0] == DATA_DESCRIPTOR_SIGNATURE:
            descriptor = read_exact(source, 12)
            record_end = data_end + 16
        else:
            descriptor = first + read_exact(source, 8)
            record_end = data_end + 12
        if struct.unpack("<III", descriptor) != (entry.CRC, entry.compress_size, entry.file_size):
            fail(f"data descriptor mismatch: {entry.filename!r}")
    else:
        if (crc, compressed_size, file_size) != (entry.CRC, entry.compress_size, entry.file_size):
            fail(f"local header sizes mismatch: {entry.filename!r}")
        record_end = data_end
    if record_end > central_offset:
        fail(f"invalid member bounds: {entry.filename!r}")
    return data_start, data_end, record_end


def add_output(data, entry, expanded_bytes, max_expanded_bytes, member_bytes, crc):
    member_bytes += len(data)
    expanded_bytes += len(data)
    if expanded_bytes > max_expanded_bytes:
        fail("expanded size exceeds limit")
    return expanded_bytes, member_bytes, binascii.crc32(data, crc) & UINT32_MAX


def stream_member(source, entry, data_start, data_end, expanded_bytes, max_expanded_bytes):
    source.seek(data_start)
    remaining = data_end - data_start
    member_bytes = 0
    crc = 0

    if entry.compress_type == zipfile.ZIP_STORED:
        while remaining:
            chunk = read_exact(source, min(CHUNK_SIZE, remaining))
            remaining -= len(chunk)
            expanded_bytes, member_bytes, crc = add_output(
                chunk, entry, expanded_bytes, max_expanded_bytes, member_bytes, crc
            )
    elif entry.compress_type == zipfile.ZIP_DEFLATED:
        decompressor = zlib.decompressobj(-zlib.MAX_WBITS)
        while remaining:
            compressed = read_exact(source, min(CHUNK_SIZE, remaining))
            remaining -= len(compressed)
            pending = compressed
            while pending:
                output = decompressor.decompress(pending, CHUNK_SIZE)
                expanded_bytes, member_bytes, crc = add_output(
                    output, entry, expanded_bytes, max_expanded_bytes, member_bytes, crc
                )
                pending = decompressor.unconsumed_tail
                if decompressor.eof:
                    if decompressor.unused_data or pending or remaining:
                        fail(f"trailing compressed data: {entry.filename!r}")
                    break
            if decompressor.eof:
                break
        if not decompressor.eof or decompressor.unused_data or decompressor.unconsumed_tail:
            fail(f"incomplete DEFLATE stream: {entry.filename!r}")
    else:
        fail(f"unsupported compression method: {entry.filename!r}")

    if member_bytes != entry.file_size:
        fail(f"declared size mismatch: {entry.filename!r}")
    if crc != entry.CRC:
        fail(f"CRC mismatch: {entry.filename!r}")
    return expanded_bytes


def inspect(paths, max_entries, max_expanded_bytes):
    entries = 0
    declared_bytes = 0
    expanded_bytes = 0
    for path in paths:
        try:
            with open(path, "rb") as source, zipfile.ZipFile(source) as archive:
                seen = set()
                records = []
                for entry in archive.infolist():
                    entries += 1
                    if entries > max_entries:
                        fail("entry count exceeds limit")
                    declared_bytes += entry.file_size
                    if declared_bytes > max_expanded_bytes:
                        fail("expanded size exceeds limit")

                    name = entry.filename
                    trimmed = name[:-1] if entry.is_dir() else name
                    if not trimmed or "\\" in trimmed or "\0" in trimmed or trimmed.startswith("/"):
                        fail(f"unsafe path: {name!r}")
                    if any(part in ("", ".", "..") for part in trimmed.split("/")):
                        fail(f"unsafe path: {name!r}")
                    if trimmed in seen:
                        fail(f"duplicate path: {name!r}")
                    seen.add(trimmed)

                    kind = stat.S_IFMT(entry.external_attr >> 16)
                    allowed = (0, stat.S_IFDIR) if entry.is_dir() else (0, stat.S_IFREG)
                    if kind not in allowed:
                        fail(f"non-regular entry: {name!r}")
                    if entry.flag_bits & 1:
                        fail(f"encrypted entry: {name!r}")
                    if entry.compress_type not in (zipfile.ZIP_STORED, zipfile.ZIP_DEFLATED):
                        fail(f"unsupported compression method: {name!r}")
                    data_start, data_end, record_end = local_record(source, entry, archive.start_dir)
                    records.append((entry, data_start, data_end, record_end))

                records.sort(key=lambda record: record[0].header_offset)
                if records and records[0][0].header_offset != 0:
                    fail("unexpected bytes before first local header")
                for index, record in enumerate(records):
                    entry, data_start, data_end, record_end = record
                    next_offset = records[index + 1][0].header_offset if index + 1 < len(records) else archive.start_dir
                    if record_end != next_offset:
                        fail(f"unexpected bytes after member: {entry.filename!r}")
                    if entry.is_dir():
                        if entry.compress_size != 0 or entry.file_size != 0 or entry.CRC != 0:
                            fail(f"directory entry carries data: {entry.filename!r}")
                        continue
                    expanded_bytes = stream_member(
                        source, entry, data_start, data_end, expanded_bytes, max_expanded_bytes
                    )
        except (EOFError, struct.error, zipfile.BadZipFile, zlib.error) as error:
            fail(f"invalid ZIP {path}: {error}")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--max-entries", required=True, type=int)
    parser.add_argument("--max-expanded-bytes", required=True, type=int)
    parser.add_argument("archives", nargs="+")
    args = parser.parse_args()
    if args.max_entries < 1 or args.max_expanded_bytes < 1:
        parser.error("limits must be positive")
    inspect(args.archives, args.max_entries, args.max_expanded_bytes)


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError) as error:
        print(f"release artifact inspection: {error}", file=sys.stderr)
        sys.exit(1)
