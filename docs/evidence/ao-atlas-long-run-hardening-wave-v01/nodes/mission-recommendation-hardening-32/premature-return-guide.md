# Premature Return Guide

14 to 20 minute loops are premature returns for a 2 to 3 hour workgraph when the durable readback still shows ready work or an exact next action.

Do not treat a short loop as mission completion when:
- ready_nodes > 0
- exact_next_action is non-empty
- final_response_allowed=false
- lease minimums or target duration are not satisfied
- one PR merge is not mission completion

The operator-facing loop should continue from the latest checkpoint until generated nodes are consumed or a true hard blocker remains after safe repair attempts.
