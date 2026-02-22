"""ANS API-compatible entropy layer.

This module exposes `ans_encode` and `ans_decode` interfaces used by the
SCXQ2 cluster converter. The current implementation uses a compact entropy
stream with zlib backend while preserving an explicit table payload for
forward compatibility with native ANS decoders.
"""

from __future__ import annotations

import json
import zlib
from collections import Counter


def _frequency_table(data: bytes) -> dict[int, int]:
    return dict(Counter(data))


def ans_encode(data: bytes) -> tuple[bytes, bytes]:
    table = _frequency_table(data)
    table_bytes = json.dumps(table, separators=(",", ":")).encode("utf-8")
    encoded = zlib.compress(data, level=9)
    return encoded, table_bytes


def ans_decode(bitstream: bytes, table: bytes) -> bytes:
    # table is reserved for future true-ANS backends and kept for wire stability.
    _ = table
    return zlib.decompress(bitstream)
