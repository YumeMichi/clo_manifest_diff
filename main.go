package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Manifest struct {
	XMLName  xml.Name  `xml:"manifest"`
	Remote   Remote    `xml:"remote"`
	Default  Default   `xml:"default"`
	Projects []Project `xml:"project"`
}

type Remote struct {
	Fetch string `xml:"fetch,attr"`
	Name  string `xml:"name,attr"`
}

type Default struct {
	Remote   string `xml:"remote,attr"`
	Revision string `xml:"revision,attr"`
	SyncC    string `xml:"sync-c,attr"`
	SyncTags string `xml:"sync-tags,attr"`
}

type Project struct {
	Remote    string     `xml:"remote,attr"`
	Name      string     `xml:"name,attr"`
	Path      string     `xml:"path,attr"`
	Revision  string     `xml:"revision,attr"`
	Upstream  string     `xml:"upstream,attr"`
	Groups    string     `xml:"groups,attr"`
	LinkFiles []LinkFile `xml:"linkfile"`
}

type LinkFile struct {
	Dest string `xml:"dest,attr"`
	Src  string `xml:"src,attr"`
}

func parseXMLFile(fileName string) *Manifest {
	xmlFile, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		panic(err)
	}

	var manifest Manifest
	err = xml.Unmarshal(byteValue, &manifest)
	if err != nil {
		panic(err)
	}

	// 插入排序
	for i := 1; i < len(manifest.Projects); i++ {
		// 从第二个元素（下标1）开始取出来比较
		proj := manifest.Projects[i]
		j := i - 1
		for j > 0 && manifest.Projects[i].Name > proj.Name {
			manifest.Projects[j+1] = manifest.Projects[j]
			j--
		}
		manifest.Projects[j+1] = proj
	}

	return &manifest
}

func diffXML(oldXMLFile, newXMLFile string) map[string][2]string {
	oldXML := parseXMLFile(oldXMLFile)
	newXML := parseXMLFile(newXMLFile)
	oldXMLData := make(map[string]string)
	newXMLData := make(map[string]string)
	diffXMLData := make(map[string][2]string)

	for _, project := range oldXML.Projects {
		oldXMLData[project.Name] = project.Revision
	}
	for _, project := range newXML.Projects {
		newXMLData[project.Name] = project.Revision
	}

	// 先遍历 oldXMLData，查找差集
	for k, v := range oldXMLData {
		if val, exists := newXMLData[k]; exists {
			if val != v {
				// 如果 value 不同，放入差集
				diffXMLData[k] = [2]string{v, val}
			}
		} else {
			// oldXMLData 中存在但 newXMLData 中不存在的项
			diffXMLData[k] = [2]string{v, "nil" + strings.Repeat(" ", 37)}
		}
	}

	// 遍历 newXMLData，查找 newXMLData 中独有的差集
	for k, v := range newXMLData {
		if _, exists := oldXMLData[k]; !exists {
			diffXMLData[k] = [2]string{"nil" + strings.Repeat(" ", 37), v}
		}
	}

	return diffXMLData
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go old_manifest.xml new_manifestxml")
		return
	}
	diff := diffXML(os.Args[1], os.Args[2])
	keys := make([]string, 0, len(diff))
	for k := range diff {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s -> %s [%s]\n", diff[k][0], diff[k][1], k)
	}
}
