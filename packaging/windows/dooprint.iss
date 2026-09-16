; Dooprint installer for Windows.
;
; Installs dooprint.exe as the "dooprint" Windows service, opens its port in the firewall and,
; at the end, opens the web interface to pair it with Odoo. Uninstalling removes the service and
; the firewall rule; the configuration in ProgramData is kept.
;
; Build (from the repository root, see the Makefile):
;   iscc /DAppVersion=1.2.0 /DBinary=..\..\dist\dooprint.exe /O..\..\dist packaging\windows\dooprint.iss

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#ifndef Binary
  #define Binary "..\..\dist\dooprint.exe"
#endif

#define AppName "Dooprint"
#define ServiceName "dooprint"
#define DefaultPort "4547"

[Setup]
AppId={{6B0E4C1F-3E0B-4B3C-9C0E-7F2D1C5A9D41}
AppName={#AppName}
AppVersion={#AppVersion}
AppVerName={#AppName} {#AppVersion}
AppPublisher=API SERVICE S.A.C
AppPublisherURL=https://github.com/apiservicesac/dooprint
AppSupportURL=https://github.com/apiservicesac/dooprint/issues
DefaultDirName={autopf}\{#AppName}
DisableProgramGroupPage=yes
DisableDirPage=auto
OutputBaseFilename=dooprint-{#AppVersion}-windows-amd64-setup
SetupIconFile=dooprint.ico
UninstallDisplayIcon={app}\dooprint.exe
UninstallDisplayName={#AppName}
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=admin
MinVersion=10.0
WizardStyle=modern
Compression=lzma2
SolidCompression=yes
VersionInfoVersion={#AppVersion}
VersionInfoProductName={#AppName}
VersionInfoDescription={#AppName} print service for Odoo

[Languages]
Name: "en"; MessagesFile: "compiler:Default.isl"
Name: "es"; MessagesFile: "compiler:Languages\Spanish.isl"

[CustomMessages]
en.PortTitle=Port
en.PortDescription=Port of the Dooprint web interface and print server
en.PortLabel=Odoo and the browsers on your network reach Dooprint on this port. Keep 4547 unless it is taken.
en.PortInvalid=Enter a port between 1 and 65535.
en.OpenInterface=Open Dooprint to pair it with Odoo
en.StartFailed=The Dooprint service was installed but did not start. Check that the port is free.
es.PortTitle=Puerto
es.PortDescription=Puerto de la interfaz web y del servidor de impresión de Dooprint
es.PortLabel=Odoo y los navegadores de tu red llegan a Dooprint por este puerto. Deja 4547 salvo que esté ocupado.
es.PortInvalid=Escribe un puerto entre 1 y 65535.
es.OpenInterface=Abrir Dooprint para emparejarlo con Odoo
es.StartFailed=El servicio de Dooprint se instaló pero no arrancó. Revisa que el puerto esté libre.

[Files]
Source: "{#Binary}"; DestDir: "{app}"; DestName: "dooprint.exe"; Flags: ignoreversion

[Dirs]
Name: "{commonappdata}\{#AppName}"

[INI]
Filename: "{app}\Dooprint.url"; Section: "InternetShortcut"; Key: "URL"; String: "http://localhost:{code:GetPort}"

[Icons]
Name: "{autoprograms}\{#AppName}"; Filename: "{app}\Dooprint.url"; IconFilename: "{app}\dooprint.exe"

[UninstallDelete]
Type: files; Name: "{app}\Dooprint.url"

[Run]
Filename: "http://localhost:{code:GetPort}"; Description: "{cm:OpenInterface}"; Flags: postinstall shellexec nowait skipifsilent

[Code]
var
  PortPage: TInputQueryWizardPage;

function GetPort(Param: String): String;
begin
  Result := Trim(PortPage.Values[0]);
end;

procedure InitializeWizard;
begin
  PortPage := CreateInputQueryPage(wpSelectDir, CustomMessage('PortTitle'), CustomMessage('PortDescription'), CustomMessage('PortLabel'));
  PortPage.Add('', False);
  PortPage.Values[0] := GetPreviousData('Port', '{#DefaultPort}');
end;

procedure RegisterPreviousData(PreviousDataKey: Integer);
begin
  SetPreviousData(PreviousDataKey, 'Port', GetPort(''));
end;

function NextButtonClick(CurPageID: Integer): Boolean;
var
  Port: Integer;
begin
  Result := True;
  if CurPageID = PortPage.ID then
  begin
    Port := StrToIntDef(GetPort(''), 0);
    if (Port < 1) or (Port > 65535) then
    begin
      MsgBox(CustomMessage('PortInvalid'), mbError, MB_OK);
      Result := False;
    end;
  end;
end;

function EscapeQuotes(S: String): String;
begin
  StringChangeEx(S, '"', '\"', True);
  Result := S;
end;

function RunHidden(const FileName, Params: String): Integer;
var
  ResultCode: Integer;
begin
  if not Exec(FileName, Params, '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then
    ResultCode := -1;
  Result := ResultCode;
end;

procedure RemoveService;
begin
  // net stop waits until the service has stopped, so the executable can be replaced.
  RunHidden(ExpandConstant('{sys}\net.exe'), 'stop {#ServiceName}');
  RunHidden(ExpandConstant('{sys}\sc.exe'), 'delete {#ServiceName}');
  RunHidden(ExpandConstant('{sys}\netsh.exe'), 'advfirewall firewall delete rule name="{#AppName}"');
end;

procedure InstallService;
var
  BinPath: String;
begin
  BinPath := '"' + ExpandConstant('{app}\dooprint.exe') + '" -port ' + GetPort('') +
    ' -data-dir "' + ExpandConstant('{commonappdata}\{#AppName}') + '"';
  RunHidden(ExpandConstant('{sys}\sc.exe'), 'create {#ServiceName} binPath= "' + EscapeQuotes(BinPath) +
    '" start= auto DisplayName= "{#AppName}"');
  RunHidden(ExpandConstant('{sys}\sc.exe'), 'description {#ServiceName} "Dooprint print service for Odoo"');
  // Start again after a failure or after the restart command sent from Odoo.
  RunHidden(ExpandConstant('{sys}\sc.exe'), 'failure {#ServiceName} reset= 86400 actions= restart/3000/restart/3000/restart/10000');
  RunHidden(ExpandConstant('{sys}\sc.exe'), 'failureflag {#ServiceName} 1');
  RunHidden(ExpandConstant('{sys}\netsh.exe'), 'advfirewall firewall add rule name="{#AppName}" dir=in action=allow protocol=TCP localport=' + GetPort(''));
  if RunHidden(ExpandConstant('{sys}\net.exe'), 'start {#ServiceName}') <> 0 then
    MsgBox(CustomMessage('StartFailed'), mbError, MB_OK);
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssInstall then
    RemoveService
  else if CurStep = ssPostInstall then
    InstallService;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usUninstall then
    RemoveService;
end;
