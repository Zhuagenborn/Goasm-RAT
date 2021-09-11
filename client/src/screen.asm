.386
.model flat, stdcall
option casemap :none

__UNICODE__     equ     1

include screen.inc
include packet.inc
include network.inc
include mem_alloc.inc

include /masm32/include/kernel32.inc
include /masm32/include/user32.inc
include /masm32/include/masm32rt.inc

includelib /masm32/lib/kernel32.lib
includelib /masm32/lib/user32.lib
includelib /masm32/lib/wsock32.lib


.data
dc          HDC         NULL
cdc         HDC         NULL
bmp_width   DWORD       0
bmp_height  DWORD       0
bmp_size    DWORD       0
bmp         HBITMAP     NULL


.code
InitDevice  proc
    invoke  CreateDC, chr$("DISPLAY"), NULL, NULL, NULL
    .if     eax == NULL
        print   "Failed to create a device context for the screen."
        jmp     _Error
    .endif
    mov     dc, eax

    invoke  CreateCompatibleDC, dc
    .if     eax == NULL
        print   "Failed to create a compatible device context for the screen."
        jmp     _Error
    .endif
    mov     cdc, eax

    invoke  GetDeviceCaps, dc, HORZRES
    mov     bmp_width, eax
    invoke  GetDeviceCaps, dc, VERTRES
    mov     bmp_height, eax
    mov     eax, bmp_width
    imul    eax, bmp_height
    imul    eax, sizeof COLORREF
    .if     edx != 0
        print   "The bitmap is too large."
        jmp     _Error
    .endif
    mov     bmp_size, eax

    invoke  CreateCompatibleBitmap, dc, bmp_width, bmp_height
    .if     eax == NULL
        print   "Failed to create a compatible bitmap for the screen."
        jmp     _Error
    .endif
    mov     bmp, eax

    invoke  SelectObject, cdc, bmp
    mov     eax, TRUE
    ret

_Error:
    invoke  FreeDevice
    mov     eax, FALSE
    ret
InitDevice  endp


FreeDevice  proc
    .if dc != NULL
        invoke  DeleteDC, dc
        mov     dc, NULL
    .endif

    .if cdc != NULL
        invoke  DeleteDC, cdc
        mov     cdc, NULL
    .endif

    .if bmp != NULL
        invoke  DeleteObject, bmp
        mov     bmp, NULL
    .endif

    ret
FreeDevice  endp


OnScreen    proc    remote: SOCKET, header: ptr Header, data: ptr BYTE
    local   @buffer: ptr BYTE
    local   @header: Header
    local   @success: BOOL

    print   "The client has received a request to take a screenshot.", 0Dh, 0Ah

    mov     @buffer, NULL
    mov     @success, FALSE
    invoke  BitBlt, cdc, 0, 0, bmp_width, bmp_height, dc, 0, 0, SRCCOPY
    .if     eax == NULL
        print   "Failed to perform a bit-block transfer from the screen to the bitmap.", 0Dh, 0Ah
        jmp     _Exit
    .endif

    invoke  RtlZeroMemory, addr @header, sizeof @header
    mov     @header.packet_type, SCREEN
    mov     eax, bmp_size
    add     eax, sizeof bmp_width + sizeof bmp_height
    mov     @header.data_size, eax

    invoke  AllocMemory, eax
    mov     @buffer, eax
    invoke  RtlMoveMemory, eax, addr bmp_width, sizeof bmp_width
    mov     eax, @buffer
    add     eax, sizeof bmp_width
    invoke  RtlMoveMemory, eax, addr bmp_height, sizeof bmp_height

    mov     eax, @buffer
    add     eax, sizeof bmp_width + sizeof bmp_height
    invoke  GetBitmapBits, bmp, bmp_size, eax
    .if     eax == 0
        print   "Failed to copy bitmap bits to the buffer.", 0Dh, 0Ah
        jmp     _Exit
    .endif

    invoke  SendPacket, remote, addr @header, @buffer
    .if     eax == SOCKET_ERROR
        invoke  OutputDebugString, uc$("Failed to send bitmap bits to the server.")
        jmp     _Exit
    .endif

    mov     @success, TRUE

_Exit:
    invoke  FreeMemory, @buffer
    mov     @buffer, NULL
    mov    eax, @success
    ret
OnScreen    endp

end