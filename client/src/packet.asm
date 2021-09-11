.386
.model flat, stdcall
option casemap :none

include packet.inc
include network.inc
include mem_alloc.inc

include /masm32/include/kernel32.inc
include /masm32/include/user32.inc

includelib /masm32/lib/kernel32.lib
includelib /masm32/lib/user32.lib
includelib /masm32/lib/wsock32.lib


.code
SendPacket  proc    remote: SOCKET, header: ptr Header, data: ptr BYTE
    local   @data_size: DWORD

    ; Send the header of the packet
    invoke  SendData, remote, header, sizeof Header
    .if     eax != SOCKET_ERROR
        mov     esi, header
        mov     eax, dword ptr [esi + Header.data_size]
        mov     @data_size, eax
        ; Send the follow-up data
        invoke  SendData, remote, data, eax
    .endif

    .if     eax != SOCKET_ERROR
        mov     eax, @data_size
        add     eax, sizeof Header
    .endif
    ret
SendPacket  endp


RecvPacket  proc    remote: SOCKET, header: ptr Header, data: ptr DWORD
    local   @data_size: DWORD
    local   @buffer: ptr BYTE

    ; Receive the header of the packet
    invoke  RecvData, remote, header, sizeof Header
    .if     eax == SOCKET_ERROR
        ret
    .endif
    mov     esi, header
    mov     eax, dword ptr [esi + Header.data_size]
    mov     @data_size, eax
    .if     eax == 0
        ; The packet contains no data
        mov     eax, sizeof Header
        ret
    .endif

    ; Receive the follow-up data
    invoke  AllocMemory, eax
    mov     @buffer, eax
    mov     ebx, data
    mov     dword ptr [ebx], eax
    invoke  RecvData, remote, eax, @data_size
    .if     eax == SOCKET_ERROR
        invoke  FreeMemory, @buffer
        mov     @buffer, NULL
        mov     eax, SOCKET_ERROR
        ret
    .endif

    mov     eax, sizeof Header
    add     eax, @data_size
    ret
RecvPacket  endp

end