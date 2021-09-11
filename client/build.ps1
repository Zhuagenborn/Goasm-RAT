New-Variable -Name 'AppName' -Value 'client' -Option Constant

$Root = (Get-Location).Path
$Bin = Join-Path -Path $Root -ChildPath 'bin'
New-Item -Name 'build' -ItemType 'Directory' -Force
New-Item -Name 'bin' -ItemType 'Directory' -Force
Set-Location -Path 'build'

$AsmFiles = Join-Path -Path $Root -ChildPath 'src' -AdditionalChildPath '*.asm'
Start-Process -FilePath 'ml' -ArgumentList '/c', '/coff', $AsmFiles -NoNewWindow -Wait
Start-Process -FilePath 'link' -ArgumentList '/subsystem:console', "/out:$AppName.exe", '*.obj' -NoNewWindow -Wait

Move-Item -Path "$AppName.exe" -Destination $Bin -Force