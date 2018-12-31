@set args=%*
@powershell "iex((@('')*3+(cat '%~f0'|select -skip 3))-join[char]10)"
@exit /b %ERRORLEVEL%

$args = @( ([regex]'"([^"]*)"').Replace($env:args,{
        $args[0].Groups[1] -replace " ",[char]1
    }) -split " " | ForEach-Object{ $_ -replace [char]1," " })

set GO "go.exe" -option constant

set CMD "Cmd" -option constant

Set-PSDebug -strict
$VerbosePreference = "Continue"

function Do-Copy($src,$dst){
    Write-Verbose "$ copy '$src' '$dst'"
    Copy-Item $src $dst
}

function Do-Remove($file){
    if( Test-Path $file ){
        Write-Verbose "$ del '$file'"
        Remove-Item $file
    }
}

function Make-Dir($folder){
    if( -not (Test-Path $folder) ){
        Write-Verbose "$ mkdir '$folder'"
        New-Item $folder -type directory | Out-Null
    }
}

Add-Type -Assembly System.Windows.Forms
function Ask-Copy($src,$dst){
    $fname = (Join-Path $dst (Split-Path $src -Leaf))
    if( Test-Path $fname ){
        if( "Yes" -ne [System.Windows.Forms.MessageBox]::Show(
            'Override "{0}" by default ?' -f $fname,
            "NYAGOS Install", "YesNo","Question","button2") ){
            return
        }
    }
    Do-Copy $src $dst
}

function ForEach-GoDir{
    Get-ChildItem . -Recurse |
    Where-Object{ $_.Extension -eq '.go' } |
    ForEach-Object{ Split-Path $_.FullName -Parent } |
    Sort-Object |
    Get-Unique
}

function Go-Fmt{
    $status = $true
    git status -s | %{
        $fname = $_.Substring(3)
        $arrow = $fname.IndexOf(" -> ")
        if( $arrow -ge 0 ){
            $fname = $fname.Substring($arrow+4)
        }
        if( $fname -like "*.go" -and (Test-Path($fname)) ){
            $prop = Get-ItemProperty($fname)
            if( $prop.Mode -like "?a*" ){
                Write-Verbose "$ $GO fmt $fname"
                & $GO fmt $fname
                if( $LastExitCode -ne 0 ){
                    $status = $false
                }else{
                    attrib -a $fname
                }
            }
        }
    }
    if( -not $status ){
        Write-Warning "Some of '$GO fmt' failed."
    }
    return $status
}

function Get-Go1stPath {
    if( $env:gopath -ne $null -and $env:gopath -ne "" ){
        $gopath = $env:gopath
    }else{
        $gopath = (Join-Path $env:userprofile "go")
    }
    $gopath.Split(";")[0]
}

function Make-SysO($version) {
    Download-Exe "github.com/josephspurrier/goversioninfo/cmd/goversioninfo" "goversioninfo.exe"
    if( $version -match "^\d+[\._]\d+[\._]\d+[\._]\d+$" ){
        $v = $version.Split("[\._]")
    }else{
        $v = @(0,0,0,0)
        if( $version -eq $null -or $version -eq "" ){
            $version = "0.0.0_0"
        }
    }
    Write-Verbose "version=$version"

    .\goversioninfo.exe `
        "-file-version=$version" `
        "-product-version=$version" `
        "-icon=nyagos.ico" `
        ("-ver-major=" + $v[0]) `
        ("-ver-minor=" + $v[1]) `
        ("-ver-patch=" + $v[2]) `
        ("-ver-build=" + $v[3]) `
        ("-product-ver-major=" + $v[0]) `
        ("-product-ver-minor=" + $v[1]) `
        ("-product-ver-patch=" + $v[2]) `
        ("-product-ver-build=" + $v[3]) `
        "-o" nyagos.syso `
        versioninfo.json
}


function Download-Exe($url,$exename){
    if( Test-Path $exename ){
        Write-Verbose -Message ("Found {0}" -f $exename)
        return
    }
    Write-Verbose -Message ("{0} not found." -f $exename)
    Write-Verbose -Message ("$ $GO get -d " + $url)
    & $GO get -d $url
    $workdir = (Join-Path (Join-Path (Get-Go1stPath) "src") $url)
    $cwd = (Get-Location)
    Set-Location $workdir
    Write-Verbose -Message ("$ $GO build {0} on {1}" -f $exename,$workdir)
    & $GO build
    Do-Copy $exename $cwd
    Set-Location $cwd
}

function Build($version,$tags) {
    Write-Verbose "Build as version='$version' tags='$tags'"

    if( -not (Go-Fmt) ){
        return
    }
    $saveGOARCH = $env:GOARCH
    $env:GOARCH = (& $go env GOARCH)

    Make-Dir $CMD
    $binDir = (Join-Path $CMD $env:GOARCH)
    Make-Dir $binDir
    $target = (Join-Path $binDir "nyagos.exe")

    Make-SysO $version

    Write-Verbose "$ $GO build -o '$target'"
    & $GO build "-o" $target -ldflags "-s -w -X main.version=$version" $tags
    if( $LastExitCode -eq 0 ){
        Do-Copy $target ".\nyagos.exe"
    }
    $env:GOARCH = $saveGOARCH
}

function Byte2DWord($a,$b,$c,$d){
    return ($a+256*($b+256*($c+256*$d)))
}

function Get-Architecture($bin){
    $addr = (Byte2DWord $bin[60] $bin[61] $bin[62] $bin[63])
    if( $bin[$addr] -eq 0x50 -and $bin[$addr+1] -eq 0x45 ){
        if( $bin[$addr+4] -eq 0x4C -and $bin[$addr+5 ] -eq 0x01 ){
            return 32
        }
        if( $bin[$addr+4] -eq 0x64 -and $bin[$addr+5] -eq 0x86 ){
            return 64
        }
    }
    return $null
}

function Make-Package($arch){
    $zipname = ("nyagos-{0}.zip" -f (& cmd\$arch\nyagos.exe --show-version-only))
    Write-Verbose "$ zip -9 $zipname ...."
    if( Test-Path $zipname ){
        Do-Remove $zipname
    }
    zip -9j $zipname `
        "cmd\$arch\nyagos.exe" `
        .nyagos `
        _nyagos `
        makeicon.cmd `
        LICENSE `
        readme_ja.md `
        readme.md

    zip -9 $zipname `
        nyagos.d\*.lua `
        nyagos.d\catalog\*.lua `
        Doc\*.md
}

switch( $args[0] ){
    "" {
        Build (git describe --tags) ""
    }
    "386"{
        $private:save = $env:GOARCH
        $env:GOARCH = "386"
        Build (git describe --tags) ""
        $env:GOARCH = $save
    }
    "debug" {
        $private:save = $env:GOARCH
        if( $args[1] ){
            $env:GOARCH = $args[1]
        }
        Build "" "-tags=debug"
        $env:GOARCH = $save
    }
    "release" {
        $private:save = $env:GOARCH
        if( $args[1] ){
            $env:GOARCH = $args[1]
        }
        Build (Get-Content Misc\version.txt) ""
        $env:GOARCH = $save
    }
    "clean" {
        foreach( $p in @(`
            (Join-Path $CMD "amd64\nyagos.exe"),`
            (Join-Path $CMD "386\nyagos.exe"),`
            "nyagos.exe",`
            "nyagos.syso",`
            "version.now",`
            "goversioninfo.exe") )
        {
            Do-Remove $p
        }
        Get-ChildItem "." -Recurse |
        Where-Object { $_.Name -eq "make.xml" } |
        ForEach-Object {
            $dir = (Split-Path $_.FullName -Parent)
            $xml = [xml](Get-Content $_.FullName)
            foreach($li in $xml.make.generate.li){
                if( -not $li ){ continue }
                foreach($target in $xml.make.generate.li.target){
                    if( -not $target ){ continue }
                    $path = (Join-Path $dir $target)
                    if( Test-Path $path ){
                        Do-Remove $path
                    }
                }
            }
        }

        ForEach-GoDir | %{
            Write-Verbose "$ $GO clean on $_"
            pushd $_
            & $GO clean
            popd
        }
    }
    "package" {
        $goarch = if( $args[1] ){ $args[1] }else{ (& $go env GOARCH) }
        Make-Package $goarch
    }
    "install" {
        $installDir = $args[1]
        if( $installDir -eq $null -or $installDir -eq "" ){
            $installDir = (
                Select-String 'INSTALLDIR=([^\)"]+)' Misc\version.cmd |
                ForEach-Object{ $_.Matches[0].Groups[1].Value }
            )
            if( -not $installDir ){
                Write-Warning "Usage: make.ps1 install INSTALLDIR"
                exit
            }
            if( -not (Test-Path $installDir) ){
                Write-Warning "$installDir not found."
                exit
            }
            Write-Verbose "installDir=$installDir"
        }
        Write-Output "@set `"INSTALLDIR=$installDir`"" |
            Out-File "Misc\version.cmd" -Encoding Default

        robocopy nyagos.d (Join-Path $installDir "nyagos.d") /E
        Write-Verbose ("ERRORLEVEL=" + $LastExitCode)
        if( $LastExitCode -lt 8 ){
            Remove-Item Variable:LastExitCode
        }
        Ask-Copy "_nyagos" $installDir
        try{
            Do-Copy nyagos.exe $installDir
        }catch{
            taskkill /F /im nyagos.exe
            try{
                Do-Copy nyagos.exe $installDir
            }catch{
                Write-Host "Could not update installed nyagos.exe"
                Write-Host "Some processes holds nyagos.exe now"
            }
            # [void]([System.Windows.Forms.MessageBox]::Show("Done"))
            timeout /T 3
        }
    }
    "get" {
        if( (git branch --contains) -eq "* master" ){
            Write-Verbose "$GO get -u -v ./..."
            & $GO get -u -v ./...
        } else {
            Write-Verbose "$GO get -v ./..."
            & $GO get -v ./...
        }
    }
    "fmt" {
        Go-Fmt | Out-Null
    }
    "help" {
        Write-Output @'
make                     build as snapshot
make debug   [386|amd64] build as debug version     (tagged as `debug`)
make release [386|amd64] build as release version
make clean               remove all work files
make package [386|amd64] make `nyagos-(VERSION)-(ARCH).zip`
make install [FOLDER]    copy executables to FOLDER or last folder
make fmt                 `go fmt`
make help                show this
'@
    }
    default {
        Write-Warning ("{0} not supported." -f $args[0])
    }
}
if( Test-Path Variable:LastExitCode ){
    exit $LastExitCode
}

# vim:set ft=ps1:
