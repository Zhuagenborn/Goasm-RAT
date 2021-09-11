include /masm32/include/windows.inc
include /masm32/include/wsock32.inc


; Type of a packet
SHELL           equ     3

StartupShell    proto   remote: SOCKET

StopShell       proto

OnShell         proto   remote: SOCKET, header: ptr Header, data: ptr BYTE