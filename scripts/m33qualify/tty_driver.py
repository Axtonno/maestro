"""Exercise the frozen approval choreography on a real Linux PTY.

Only disposable qualification fixtures are approved. This is not a user-facing
approval adapter. The Go harness refuses an allow on an unexpected candidate.
"""
import os
import pty
import re
import select
import sys
import time

pid, fd = pty.fork()
if pid == 0:
    os.execv(sys.argv[1], sys.argv[1:])

pending = b""
choice = None
last_output = time.monotonic()
try:
    while True:
        ready, _, _ = select.select([fd], [], [], 10)
        if not ready:
            if time.monotonic() - last_output > 420:
                raise TimeoutError("qualification stalled")
            continue
        try:
            chunk = os.read(fd, 65536)
        except OSError:
            break
        if not chunk:
            break
        last_output = time.monotonic()
        sys.stdout.buffer.write(chunk)
        sys.stdout.buffer.flush()
        pending += chunk
        match = re.search(rb"M33 approval task=\S+ expected=(allow|deny)\r?\n", pending)
        if match:
            choice = b"o\n" if match[1] == b"allow" else b"d\n"
            pending = pending[match.end():]
        prompt = b"(default deny): "
        if prompt in pending:
            if choice is None:
                raise RuntimeError("unplanned approval")
            os.write(fd, choice)
            choice = None
            pending = pending.split(prompt, 1)[1]
        if len(pending) > 262144:
            pending = pending[-65536:]
finally:
    os.close(fd)
_, status = os.waitpid(pid, 0)
sys.exit(os.waitstatus_to_exitcode(status))
