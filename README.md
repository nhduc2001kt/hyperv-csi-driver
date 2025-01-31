# Hyper-V CSI Driver

The Hyper-V CSI Driver, implemented in Go, is a Container Storage Interface (CSI) driver designed for Kubernetes to integrate and manage virtual disks on the Microsoft Hyper-V platform.

## In Scope

* [Dynamic provisioning](https://kubernetes-csi.github.io/docs/external-provisioner.html): Volumes are created dynamically when `PersistentVolumeClaim` objects are created.
* [Static provisioning](https://kubernetes-csi.github.io/docs/external-provisioner.html): Volumes are manually provisioned by administrators and referenced by `PersistentVolume` objects.
* [Volume metrics] (not yet): Usage stats are exported as Prometheus metrics from `kubelet`.
* [Volume expansion](https://kubernetes-csi.github.io/docs/volume-expansion.html) (not yet): Volumes can be expanded by editing `PersistentVolumeClaim` objects.
* [Storage capacity](https://kubernetes.io/docs/concepts/storage/storage-capacity/) (not yet): Kubernetes supports storage capacity tracking to ensure that volume provisioning respects the available storage in the cluster.


## Prerequisite
Setting up WinRM for Driver usage:
* Enable WinRM with negotiate authentication support
```
Enable-PSRemoting -SkipNetworkProfileCheck -Force

Set-WSManInstance WinRM/Config/WinRS -ValueSet @{MaxMemoryPerShellMB = 1024}
Set-WSManInstance WinRM/Config -ValueSet @{MaxTimeoutms=1800000}
Set-WSManInstance WinRM/Config/Client -ValueSet @{TrustedHosts="*"}
Set-WSManInstance WinRM/Config/Service/Auth -ValueSet @{Negotiate = $true}
```
* WinRM allow HTTPS
```
#Create CA certificate
$rootCaName = "DevRootCA"
$rootCaPassword = ConvertTo-SecureString "P@ssw0rd" -asplaintext -force 
$rootCaCertificate = Get-ChildItem cert:\LocalMachine\Root |?{$_.subject -eq "CN=$rootCaName"}
if (!$rootCaCertificate){
  Get-ChildItem cert:\LocalMachine\My |?{$_.subject -eq "CN=$rootCaName"} | remove-item -force
  if (Test-Path .\$rootCaName.cer) {
    remove-item .\$rootCaName.cer -force
  }
  if (Test-Path .\$rootCaName.pfx) {
    remove-item .\$rootCaName.pfx -force
  }
  $params = @{
    Type = 'Custom'
    DnsName = $rootCaName
    Subject = "CN=$rootCaName"
    KeyExportPolicy = 'Exportable'
    CertStoreLocation = 'Cert:\LocalMachine\My'
    KeyUsageProperty = 'All'
    KeyUsage = 'None'
    Provider = 'Microsoft Strong Cryptographic Provider'
    KeySpec = 'KeyExchange'
    KeyLength = 4096
    HashAlgorithm = 'SHA256'
    KeyAlgorithm = 'RSA'
    NotAfter = (Get-Date).AddYears(5)
  }
  $rootCaCertificate = New-SelfSignedCertificate @params

  Export-Certificate -Cert $rootCaCertificate -FilePath .\$rootCaName.cer -Verbose
  Export-PfxCertificate -Cert $rootCaCertificate -FilePath .\$rootCaName.pfx -Password $rootCaPassword -Verbose
  Get-ChildItem cert:\LocalMachine\My |?{$_.subject -eq "CN=$rootCaName"} | remove-item -force
  Import-PfxCertificate -FilePath .\$rootCaName.pfx -CertStoreLocation Cert:\LocalMachine\Root -password $rootCaPassword -Exportable -Verbose
  Import-PfxCertificate -FilePath .\$rootCaName.pfx -CertStoreLocation Cert:\LocalMachine\My -password $rootCaPassword -Exportable -Verbose
  $rootCaCertificate = Get-ChildItem cert:\LocalMachine\My |?{$_.subject -eq "CN=$rootCaName"}
}

#Create host certificate using CA
$hostName = [System.Net.Dns]::GetHostName()
$hostPassword = ConvertTo-SecureString "P@ssw0rd" -asplaintext -force
$hostCertificate = Get-ChildItem cert:\LocalMachine\My |?{$_.subject -eq "CN=$hostName"}
if (!$hostCertificate){
  if (Test-Path .\$hostName.cer) {
    remove-item .\$hostName.cer -force
  }
  if (Test-Path .\$hostName.pfx) {
    remove-item .\$hostName.pfx -force
  }
  $dnsNames = @($hostName, "localhost", "127.0.0.1") + [System.Net.Dns]::GetHostByName($env:computerName).AddressList.IpAddressToString
  
  $params = @{
    Type = 'Custom'
    DnsName = $dnsNames
    Subject = "CN=$hostName"
    KeyExportPolicy = 'Exportable'
    CertStoreLocation = 'Cert:\LocalMachine\My'
    KeyUsageProperty = 'All'
    KeyUsage = @('KeyEncipherment','DigitalSignature','NonRepudiation')
    TextExtension = @("2.5.29.37={text}1.3.6.1.5.5.7.3.1,1.3.6.1.5.5.7.3.2")
    Signer = $rootCaCertificate
    Provider = 'Microsoft Strong Cryptographic Provider'
    KeySpec = 'KeyExchange'
    KeyLength = 2048
    HashAlgorithm = 'SHA256'
    KeyAlgorithm = 'RSA'
    NotAfter = (Get-date).AddYears(2)
  }
  $hostCertificate = New-SelfSignedCertificate @params
  Export-Certificate -Cert $hostCertificate -FilePath .\$hostName.cer -Verbose
  Export-PfxCertificate -Cert $hostCertificate -FilePath .\$hostName.pfx -Password $hostPassword -Verbose
  Get-ChildItem cert:\LocalMachine\My |?{$_.subject -eq "CN=$hostName"} | remove-item -force
  Import-PfxCertificate -FilePath .\$hostName.pfx -CertStoreLocation Cert:\LocalMachine\My -password $hostPassword -Exportable -Verbose
  $hostCertificate = Get-ChildItem cert:\LocalMachine\My |?{$_.subject -eq "CN=$hostName"}
}

Get-ChildItem wsman:\localhost\Listener\ | Where-Object -Property Keys -eq 'Transport=HTTPS' | Remove-Item -Recurse
New-Item -Path WSMan:\localhost\Listener -Transport HTTPS -Address * -CertificateThumbPrint $($hostCertificate.Thumbprint) -Force -Verbose

Restart-Service WinRM -Verbose

New-NetFirewallRule -DisplayName "Windows Remote Management (HTTPS-In)" -Name "WinRMHTTPSIn" -Profile Any -LocalPort 5986 -Protocol TCP -Verbose
```
* WinRM allow HTTP
```
# Get the public networks
$PubNets = Get-NetConnectionProfile -NetworkCategory Public -ErrorAction SilentlyContinue 

# Set the profile to private
foreach ($PubNet in $PubNets) {
    Set-NetConnectionProfile -InterfaceIndex $PubNet.InterfaceIndex -NetworkCategory Private
}

# Configure winrm
Set-WSManInstance WinRM/Config/Service -ValueSet @{AllowUnencrypted = $true}

# Restore network categories
foreach ($PubNet in $PubNets) {
    Set-NetConnectionProfile -InterfaceIndex $PubNet.InterfaceIndex -NetworkCategory Public
}

Get-ChildItem wsman:\localhost\Listener\ | Where-Object -Property Keys -eq 'Transport=HTTP' | Remove-Item -Recurse
New-Item -Path WSMan:\localhost\Listener -Transport HTTP -Address * -Force -Verbose

Restart-Service WinRM -Verbose

New-NetFirewallRule -DisplayName "Windows Remote Management (HTTP-In)" -Name "WinRMHTTPIn" -Profile Any -LocalPort 5985 -Protocol TCP -Verbose
```

## Installation
### Kustomize
Give your own username and password in the file `./deploy/kubernetes/overlays/latest/kustomization.yaml`
Then, run this command
```sh
kubectl apply -k "./deploy/kubernetes/overlays/latest"
```

## Example
### Static provisioning
Manually create a vhdx file on Windows host with this Powershell script:
```
$ErrorActionPreference = 'Stop'

Import-Module Hyper-V

$NewVHDArgs = @{
  Path  = "C:\ProgramData\Microsoft\Windows\Virtual Hard Disks\test.vhdx"
  Size  = 1073741824 # 1GB
  Fixed = $true
}

New-VHD @NewVHDArgs
```
Then, run this command
```sh
kubectl apply -f "./examples/static-provisioning/manifests"
```
### Dynamic provisioning
Run this command
```sh
kubectl apply -f "./examples/dynamic-provisioning/manifests"
```