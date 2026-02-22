from __future__ import annotations

import hashlib
from typing import Sequence


def hash_node(data: bytes) -> bytes:
    return hashlib.sha256(data).digest()


def merkle_root(nodes: Sequence[bytes]) -> bytes:
    if not nodes:
        return hash_node(b"")

    hashes = [hash_node(n) for n in nodes]
    while len(hashes) > 1:
        nxt: list[bytes] = []
        for i in range(0, len(hashes), 2):
            a = hashes[i]
            b = hashes[i + 1] if i + 1 < len(hashes) else a
            nxt.append(hash_node(a + b))
        hashes = nxt

    return hashes[0]
