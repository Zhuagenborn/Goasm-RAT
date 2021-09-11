include /masm32/include/windows.inc
include /masm32/include/wsock32.inc


InitWinSocket   proto

FreeWinSocket   proto

RecvData        proto   remote: SOCKET, data: ptr BYTE, data_size: DWORD

SendData        proto   remote: SOCKET, data: ptr BYTE, data_size: DWORD

StartupService  proto   ip: ptr BYTE, port: DWORD

StopService     proto