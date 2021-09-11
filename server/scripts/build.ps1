New-Variable -Name 'AppName' -Value 'server' -Option Constant

Set-Location -Path '..'
$Root = (Get-Location).Path
New-Item -Name 'build' -ItemType 'Directory' -Force

Set-Location -Path (Join-Path -Path 'cmd' -ChildPath $AppName)
Start-Process -FilePath 'go' -ArgumentList 'build', "-o $AppName.exe" -NoNewWindow -Wait
Move-Item -Path "$AppName.exe" -Destination (Join-Path -Path $Root -ChildPath 'build') -Force