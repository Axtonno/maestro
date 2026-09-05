"""Drive only the frozen M35 disposable-fixture approval choreography."""
import os, pty, re, select, sys, time
pid, fd = pty.fork()
if pid == 0: os.execv(sys.argv[1], sys.argv[1:])
pending=b""; choice=None; last=time.monotonic()
try:
    while True:
        ready,_,_=select.select([fd],[],[],10)
        if not ready:
            if time.monotonic()-last>420: raise TimeoutError("qualification stalled")
            continue
        try: chunk=os.read(fd,65536)
        except OSError: break
        if not chunk: break
        last=time.monotonic();sys.stdout.buffer.write(chunk);sys.stdout.buffer.flush();pending+=chunk
        match=re.search(rb"M35 approval task=\S+ expected=(allow|deny)\r?\n",pending)
        if match: choice=b"o\n" if match[1]==b"allow" else b"d\n";pending=pending[match.end():]
        marker=b"(default deny): "
        if marker in pending:
            if choice is None: raise RuntimeError("unplanned approval")
            os.write(fd,choice);choice=None;pending=pending.split(marker,1)[1]
        if len(pending)>262144:pending=pending[-65536:]
finally: os.close(fd)
_,status=os.waitpid(pid,0);sys.exit(os.waitstatus_to_exitcode(status))
