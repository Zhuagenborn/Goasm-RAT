include /masm32/include/windows.inc
include /masm32/include/wsock32.inc


; Type of a packet
SCREEN      equ     4

InitDevice  proto

FreeDevice  proto

OnScreen    proto   remote: SOCKET, header: ptr Header, data: ptr BYTE