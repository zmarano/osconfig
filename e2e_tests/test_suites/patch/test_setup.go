//  Copyright 2019 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package patch

import (
	"time"

	"github.com/GoogleCloudPlatform/osconfig/e2e_tests/compute"
	"github.com/GoogleCloudPlatform/osconfig/e2e_tests/utils"
	computeApi "google.golang.org/api/compute/v1"
)

type patchTestSetup struct {
	testName      string
	image         string
	metadata      []*computeApi.MetadataItems
	assertTimeout time.Duration
	machineType   string
}

var (
	windowsRecordBoot = `
while ($true) {
  $uri = 'http://metadata.google.internal/computeMetadata/v1/instance/guest-attributes/osconfig_tests/boot_count'
  $old = Invoke-RestMethod -Method GET -Uri $uri -Headers @{"Metadata-Flavor" = "Google"}
  $new = $old+1
  try {
    Invoke-RestMethod -Method PUT -Uri $uri -Headers @{"Metadata-Flavor" = "Google"} -Body $new -ErrorAction Stop
  }
  catch {
    Write-Output $_.Exception.Message
    Start-Sleep 1
    continue
  }
  break
}
`
	windowsSetWsus = `
$wu_server = '192.168.0.2'
$windows_update_path = 'HKLM:\SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate'
$windows_update_au_path = "$windows_update_path\AU"

if (Test-Connection $wu_server -Count 1 -ErrorAction SilentlyContinue) {
	if (-not (Test-Path $windows_update_path -ErrorAction SilentlyContinue)) {
		New-Item -Path $windows_update_path -Value ""
		New-Item -Path $windows_update_au_path -Value ""
	}
	Set-ItemProperty -Path $windows_update_path -Name WUServer -Value "http://${wu_server}:8530"
	Set-ItemProperty -Path $windows_update_path -Name WUStatusServer -Value "http://${wu_server}:8530"
	Set-ItemProperty -Path $windows_update_au_path -Name UseWUServer -Value 1
}
`

	windowsLocalPostPatchScript = `
$uri = 'http://metadata.google.internal/computeMetadata/v1/instance/guest-attributes/osconfig_tests/post_step_ran'
New-Item -Path . -Name "windows_local_post_patch_script.ps1" -ItemType "file" -Value "Invoke-RestMethod -Method PUT -Uri $uri -Headers @{'Metadata-Flavor' = 'Google'} -Body 1"
`

	linuxRecordBoot = `
uri=http://metadata.google.internal/computeMetadata/v1/instance/guest-attributes/osconfig_tests/boot_count
old=$(curl $uri -H "Metadata-Flavor: Google" -f)
new=$(($old + 1))
curl -X PUT --data "${new}" $uri -H "Metadata-Flavor: Google"
`

	linuxLocalPrePatchScript = `
echo 'curl -X PUT --data "1" http://metadata.google.internal/computeMetadata/v1/instance/guest-attributes/osconfig_tests/pre_step_ran -H "Metadata-Flavor: Google"' >> ./linux_local_pre_patch_script.sh
chmod +x ./linux_local_pre_patch_script.sh
`

	setUpDowngradeState = `
echo 'deb [trusted=yes check-valid-until=no] http://snapshot.debian.org/archive/debian/20190801T025637Z/ buster main' >> /etc/apt/sources.list
echo 'Package: sudo' >> /etc/apt/preferences
echo 'Pin: version 1.8.27-1' >> /etc/apt/preferences
echo 'Pin-priority: 9999' >> /etc/apt/preferences
`

	enableOsconfig  = compute.BuildInstanceMetadataItem("enable-osconfig", "true")
	disableFeatures = compute.BuildInstanceMetadataItem("osconfig-disabled-features", "guestpolicies,osinventory")

	windowsSetup = &patchTestSetup{
		assertTimeout: 60 * time.Minute,
		metadata: []*computeApi.MetadataItems{
			compute.BuildInstanceMetadataItem("sysprep-specialize-script-ps1", windowsSetWsus),
			compute.BuildInstanceMetadataItem("windows-startup-script-ps1", windowsRecordBoot+utils.InstallOSConfigGooGet()+windowsLocalPostPatchScript),
			enableOsconfig,
			disableFeatures,
		},
		machineType: "e2-standard-4",
	}
	aptSetup = &patchTestSetup{
		assertTimeout: 10 * time.Minute,
		metadata: []*computeApi.MetadataItems{
			compute.BuildInstanceMetadataItem("startup-script", linuxRecordBoot+utils.InstallOSConfigDeb()+linuxLocalPrePatchScript),
			enableOsconfig,
			disableFeatures,
		},
		machineType: "e2-medium",
	}
	aptDowngradeSetup = &patchTestSetup{
		assertTimeout: 10 * time.Minute,
		metadata: []*computeApi.MetadataItems{
			compute.BuildInstanceMetadataItem("startup-script", linuxRecordBoot+utils.InstallOSConfigDeb()+linuxLocalPrePatchScript+setUpDowngradeState),
			enableOsconfig,
			disableFeatures,
		},
		machineType: "e2-medium",
	}
	el7Setup = &patchTestSetup{
		assertTimeout: 15 * time.Minute,
		metadata: []*computeApi.MetadataItems{
			compute.BuildInstanceMetadataItem("startup-script", linuxRecordBoot+utils.InstallOSConfigEL7()+linuxLocalPrePatchScript),
			enableOsconfig,
			disableFeatures,
		},
		machineType: "e2-medium",
	}
	el8Setup = &patchTestSetup{
		assertTimeout: 15 * time.Minute,
		metadata: []*computeApi.MetadataItems{
			compute.BuildInstanceMetadataItem("startup-script", linuxRecordBoot+utils.InstallOSConfigEL8()+linuxLocalPrePatchScript),
			enableOsconfig,
			disableFeatures,
		},
		machineType: "e2-medium",
	}
	el9Setup = &patchTestSetup{
		assertTimeout: 15 * time.Minute,
		metadata: []*computeApi.MetadataItems{
			compute.BuildInstanceMetadataItem("startup-script", linuxRecordBoot+utils.InstallOSConfigEL9()+linuxLocalPrePatchScript),
			enableOsconfig,
			disableFeatures,
		},
		machineType: "e2-medium",
	}
	suseSetup = &patchTestSetup{
		assertTimeout: 15 * time.Minute,
		metadata: []*computeApi.MetadataItems{
			compute.BuildInstanceMetadataItem("startup-script", linuxRecordBoot+utils.InstallOSConfigSUSE()+linuxLocalPrePatchScript),
			enableOsconfig,
			disableFeatures,
		},
		machineType: "e2-medium",
	}
)

func imageTestSetup(mapping map[*patchTestSetup]map[string]string) (setup []*patchTestSetup) {
	for s, m := range mapping {
		for name, image := range m {
			new := patchTestSetup(*s)
			new.testName = name
			new.image = image
			setup = append(setup, &new)
		}
	}
	return
}

func headImageTestSetup() []*patchTestSetup {
	// This maps a specific patchTestSetup to test setup names and associated images.
	mapping := map[*patchTestSetup]map[string]string{
		windowsSetup: utils.HeadWindowsImages,
		el7Setup:     utils.HeadEL7Images,
		el8Setup:     utils.HeadEL8Images,
		el9Setup:     utils.HeadEL9Images,
		aptSetup:     utils.HeadAptImages,
		suseSetup:    utils.HeadSUSEImages,
	}

	return imageTestSetup(mapping)
}

func oldImageTestSetup() []*patchTestSetup {
	// This maps a specific patchTestSetup to test setup names and associated images.
	mapping := map[*patchTestSetup]map[string]string{
		windowsSetup: utils.OldWindowsImages,
		el7Setup:     utils.OldEL7Images,
		el8Setup:     utils.OldEL8Images,
		el9Setup:     utils.OldEL9Images,
		aptSetup:     utils.OldAptImages,
		suseSetup:    utils.OldSUSEImages,
	}

	return imageTestSetup(mapping)
}

func aptHeadImageTestSetup() []*patchTestSetup {
	// This maps a specific patchTestSetup to test setup names and associated images.
	mapping := map[*patchTestSetup]map[string]string{
		aptSetup: utils.HeadAptImages,
	}

	return imageTestSetup(mapping)
}

func aptDowngradeImageTestSetup() []*patchTestSetup {
	// This maps a specific patchTestSetup to test setup names and associated images.
	mapping := map[*patchTestSetup]map[string]string{
		aptDowngradeSetup: utils.DowngradeAptImages,
	}

	return imageTestSetup(mapping)
}

func yumHeadImageTestSetup() []*patchTestSetup {
	// This maps a specific patchTestSetup to test setup names and associated images.
	mapping := map[*patchTestSetup]map[string]string{
		el7Setup: utils.HeadEL7Images,
		el8Setup: utils.HeadEL8Images,
		el9Setup: utils.HeadEL9Images,
	}

	return imageTestSetup(mapping)
}

func suseHeadImageTestSetup() []*patchTestSetup {
	// This maps a specific patchTestSetup to test setup names and associated images.
	mapping := map[*patchTestSetup]map[string]string{
		suseSetup: utils.HeadSUSEImages,
	}

	return imageTestSetup(mapping)
}
