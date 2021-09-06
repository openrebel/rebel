$build = 0
if (Test-Path ".\.buildcount" -PathType Leaf) { #exist
    $build = ([int](Get-Content .\.buildcount)) +1
    Out-File -FilePath .\.buildcount -InputObject $build
}

$appname = "rebel"
$version = "1.0.$($build)"
$isRelease = $args.Contains("--release")

$targetArch = [string]@(
    #"386", "amd64", "arm"
    "amd64"
)

$Env:GOOS = "windows"

foreach ($arch in $targetArch) {
    $Env:GOARCH = $arch
    $filename = "$($appname)-win-$($arch).exe"

    if (Test-Path $filename -PathType Leaf) { #exist
        Remove-Item .\$filename
    }
    
    if ($isRelease) {
        go build `
            -o ./$filename `
            -ldflags "-X main.product_release=true -X main.product_version=$($version) -s -w" `
            .\..
    } else {
        go build `
            -o ./$filename `
            -ldflags "-X main.product_release=false -X main.product_version=$($version)" `
            .\..
    }
}

if ($args.Contains("--run") || $args.Contains("--run64")) {
    #Start-Process .\$($appname)-win-amd64.exe -WorkingDirectory .\
    .\rebel-win-amd64.exe
} elseif ($args.Contains("--run32")) {
    #Start-Process .\$($appname)-win-386.exe -WorkingDirectory .\
    .\rebel-win-386.exe
}
