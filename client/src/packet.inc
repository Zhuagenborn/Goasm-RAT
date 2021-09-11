include /masm32/include/windows.inc
include /masm32/include/wsock32.inc


; Type of a packet
UNKNOW      equ     0


; Header of a packet
Header      struct
    packet_type     DWORD   UNKNOW
    data_size       DWORD   0
Header      ends


SendPacket  proto   remote: SOCKET, header: ptr Header, data: ptr BYTE

RecvPacket  proto   remote: SOCKET, header: ptr Header, data: ptr DWORD