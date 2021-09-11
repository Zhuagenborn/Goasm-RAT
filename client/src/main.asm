.386
.model flat, stdcall
option casemap :none

__UNICODE__     equ     1

include network.inc

include /masm32/include/windows.inc
include /masm32/include/shlwapi.inc
include /masm32/include/kernel32.inc
include /masm32/include/user32.inc
include /masm32/include/masm32rt.inc

includelib /masm32/lib/kernel32.lib
includelib /masm32/lib/user32.lib
includelib /masm32/lib/shlwapi.lib


STIF_DEFAULT    equ     0


.data
argc        DWORD       0
argv        DWORD       NULL
uip         DWORD       NULL
port        DWORD       0
ip          BYTE        16 dup(0)


.code
start:
    invoke  GetCommandLineW
    invoke  CommandLineToArgvW, eax, addr argc
    mov     argv, eax
    .if     argc <= 2
        print   "Usage: client <ipv4-addr> <port>", 0Dh, 0Ah
        ret
    .endif

    mov     ebx, argv
    add     ebx, sizeof DWORD
    mov     eax, dword ptr [ebx]
    mov     uip, eax
    invoke  WideCharToMultiByte, CP_ACP, 0, uip, -1, addr ip, sizeof ip, NULL, 0
    .if     eax == 0
        print   "Usage: client <ipv4-addr> <port>", 0Dh, 0Ah
        ret
    .endif

    mov     ebx, argv
    add     ebx, 2 * sizeof DWORD
    mov     eax, dword ptr [ebx]
    invoke  StrToIntEx, eax, STIF_DEFAULT, addr port
    .if     eax != TRUE
        print   "Usage: client <ipv4-addr> <port>", 0Dh, 0Ah
        ret
    .endif

    print   "The client has started.", 0Dh, 0Ah
    invoke  InitWinSocket
    .if     eax == TRUE
        invoke  StartupService, addr ip, port
        print   "The client has exited.", 0Dh, 0Ah
        invoke  FreeWinSocket
    .endif
    ret

end start