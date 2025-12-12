package utils

import (
	"log"
	"runtime"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"periph.io/x/host/v3/distro"
)

// IsLinux checks if the current operating system is Linux.
func IsLinux() bool {
	switch runtime.GOOS {
	case "linux":
		return true
	default:
		return false
	}
}

// GetLinuxDistId retrieves the distribution ID of the current Linux system.
//
// @return The distribution ID of the current Linux system, or an empty string if not found.
func GetLinuxDistID() string {
	m := distro.OSRelease()
	distID, found := m["ID"]
	if found {
		return distID
	}
	return ""
}

// IsUbuntu checks if the current Linux distribution is Ubuntu.
//
// @return true if the current OS is Ubuntu, false otherwise.
func IsUbuntu() bool {
	return GetLinuxDistID() == "ubuntu"
}

// IsDebian checks if the current Linux distribution is Debian.
//
// @return true if the current OS is Debian, false otherwise.
func IsDebian() bool {
	return GetLinuxDistID() == "debian"
}

// IsDebianSeries checks if the current Linux distribution belongs to the Debian series.
//
// @return true if the current OS is part of the Debian series, false otherwise.
func IsDebianSeries() bool {
	ids := []string{
		"debian",
		"antix",
		"lmde",
	}
	distID := GetLinuxDistID()
	for _, id := range ids {
		if distID == id {
			return true
		}
	}
	return false
}

// IsDebianUbuntuSeries checks if the current Linux distribution belongs to the Debian or Ubuntu series.
//
// @return true if the current OS is part of the Debian or Ubuntu series, false otherwise.
func IsDebianUbuntuSeries() bool {
	ids := []string{
		"debian",
		"antix",
		"lmde",
		"ubuntu", "linuxmint", "pop",
	}
	distID := GetLinuxDistID()
	for _, id := range ids {
		if distID == id {
			return true
		}
	}
	return false
}

// IsUbuntuSeries checks if the current Linux distribution belongs to the Ubuntu series.
//
// @return true if the current OS is part of the Ubuntu series, false otherwise.
func IsUbuntuSeries() bool {
	ids := []string{
		"ubuntu", "linuxmint", "pop",
	}
	distId := GetLinuxDistID()
	for _, id := range ids {
		if distId == id {
			return true
		}
	}
	return false
}

// IsFedoraSeries checks if the current Linux distribution belongs to the Fedora series.
//
// @return true if the current OS is part of the Fedora series, false otherwise.
func IsFedoraSeries() bool {
	ids := []string{
		"fedora", "centos", "rhel",
	}
	distId := GetLinuxDistID()
	for _, id := range ids {
		if distId == id {
			return true
		}
	}
	return false
}

// BuildKernelOSKeywords constructs a list of keywords based on kernel architecture and operating system.
//
// @param keywords A map where keys are keyword categories and values are lists of keywords.
//
// @return A slice of strings representing the combined list of keywords.
//
// @example
//
//	keywords := map[string][]string{
//		"common":             {"keyword1", "keyword2"},
//		"amd64":             {"amd64_keyword"},
//		"arm64":              {"arm64_keyword"},
//		"darwin":             {"darwin_keyword"},
//		"linux":              {"linux_keyword"},
//		"DebianUbuntuSeries": {"debian_ubuntu_keyword"},
//		"FedoraSeries":       {"fedora_keyword"},
//		"OtherLinux":         {"other_linux_keyword"},
//	}
//	result := BuildKernelOSKeywords(keywords)
//	// result might contain a combination of the above keywords based on the OS and architecture

func BuildKernelOSKeywords(keywords map[string][]string) []string {
	kwds := keywords["common"]
	k, found := keywords[HostKernelArch()]
	if found {
		kwds = append(kwds, k...)
	}
	k, found = keywords[runtime.GOOS]
	if found {
		kwds = append(kwds, k...)
	}
	if IsDebianUbuntuSeries() {
		debianUbuntuSeries, found := keywords["DebianUbuntuSeries"]
		if found {
			kwds = append(kwds, debianUbuntuSeries...)
		}
	} else if IsFedoraSeries() {
		fedoraSeries, found := keywords["FedoraSeries"]
		if found {
			kwds = append(kwds, fedoraSeries...)
		}
	} else {
		otherLinux, found := keywords["OtherLinux"]
		if found {
			kwds = append(kwds, otherLinux...)
		}
	}
	return kwds
}

func HostInfo() *host.InfoStat {
	info, err := host.Info()
	if err != nil {
		log.Fatal(err)
	}
	return info
}

func HostKernelArch() string {
	switch HostInfo().KernelArch {
	case "x86_64", "amd64":
		return "amd64"
	case "arm64", "aarch64":
		return "arm64"
	default:
		return "_other"
	}
}

// VirtualMemory retrieves information about the system's virtual memory.
//
// @return A pointer to a `mem.VirtualMemoryStat` struct representing the system's virtual memory information.
func VirtualMemory() *mem.VirtualMemoryStat {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return memStat
}

// CpuInfo retrieves information about the system's CPU.
//
// @return A slice of `cpu.InfoStat` structs, each representing information about a logical CPU core.
func CpuInfo() []cpu.InfoStat {
	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return cpuInfo
}
