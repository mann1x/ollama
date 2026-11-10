#!/usr/bin/env python3
"""Cut upstream's MLX payload out of ollama-darwin.tgz, entry for entry.

Members are copied with their headers (mode, mtime, owner, pax xattrs) and
their bytes unchanged, so every file's sha256 equals upstream's.
usage: make-darwin-mlx.py <ollama-darwin.tgz> <out.tgz>
"""
import gzip
import hashlib
import io
import sys
import tarfile

KEEP_DIRS = ("mlx_metal_v3", "mlx_metal_v4")
KEEP_FILES = {"MLX_LICENSE", "MLX_C_LICENSE", "XGRAMMAR_LICENSE", "XGRAMMAR_NOTICE",
              "DLPACK_LICENSE", "PICOJSON_LICENSE", "FMT_LICENSE"}

def wanted(name):
    name = name.lstrip("./")
    return name in KEEP_FILES or any(name == d or name.startswith(d + "/") for d in KEEP_DIRS)

src, out = sys.argv[1], sys.argv[2]
rows = []
with tarfile.open(src, "r:gz") as tin, open(out, "wb") as raw:
    with gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as gz:
        with tarfile.open(fileobj=gz, mode="w", format=tarfile.PAX_FORMAT) as tout:
            for m in tin:
                if not wanted(m.name):
                    continue
                if m.isreg():
                    data = tin.extractfile(m).read()
                    rows.append((hashlib.sha256(data).hexdigest(), m.size, m.name))
                    tout.addfile(m, io.BytesIO(data))
                else:
                    tout.addfile(m)
for sha, size, name in rows:
    print(f"{sha}  {size:>10}  {name}")
