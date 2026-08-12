#!/usr/bin/env python3
import json
import sys
from pathlib import Path


def object_without_duplicates(pairs):
    value = {}
    for key, item in pairs:
        if key in value:
            raise ValueError(f"duplicate key: {key}")
        value[key] = item
    return value


def reject_constant(value):
    raise ValueError(f"invalid constant: {value}")


def validate(path):
    with Path(path).open(encoding="utf-8") as source:
        value = json.load(
            source,
            object_pairs_hook=object_without_duplicates,
            parse_constant=reject_constant,
        )
    if not isinstance(value, dict):
        raise ValueError("root must be an object")


for filename in sys.argv[1:]:
    try:
        validate(filename)
    except (OSError, UnicodeError, json.JSONDecodeError, ValueError) as error:
        raise SystemExit(f"{filename}: {error}")

if len(sys.argv) == 1:
    raise SystemExit("usage: validate-strict-json.py FILE [FILE ...]")
