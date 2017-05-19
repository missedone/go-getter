package getter

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// MvnGetter is a Getter implementation that will download an artifact from maven repository, e.g. Sonatype Nexus,
// uri format: mvn::http://[username@]hostname[:port]/directoryname[?options]
type MvnGetter struct {
	HttpGet HttpGetter
}

func (g *MvnGetter) ClientMode(u *url.URL) (ClientMode, error) {
	return ClientModeFile, nil
}

func (g *MvnGetter) Get(dst string, u *url.URL) error {
	return fmt.Errorf("MvnGetter does not support download folder.")
}

// Get the remote file.
// If the version is a snapshot version, it will get the latest snapshot artifact.
// Query parameters:
//   - groupId: the group id
//   - artifactId: the artifact id
//   - version: the artifact version
//   - type: the artifact type, default as 'jar'
// example url: mvn::http://username@host/mavan/repo/path?groupId=org.example&artifactId=test&version=1.0.0-SNAPSHOT
func (g *MvnGetter) GetFile(dst string, u *url.URL) error {
	groupId := u.Query().Get("groupId")
	if groupId == "" {
		return fmt.Errorf("query parameter 'groupId' is required.")
	}
	artifactId := u.Query().Get("artifactId")
	if artifactId == "" {
		return fmt.Errorf("query parameter 'artifactId' is required.")
	}
	version := u.Query().Get("version")
	if version == "" {
		return fmt.Errorf("query parameter 'version' is required.")
	}
	artType := u.Query().Get("type")
	if artType == "" {
		artType = "jar"
	}

	artifactUrl, err := url.Parse(u.String())
	if err != nil {
		return err
	}
	artifactUrl.RawQuery = ""
	artifactUrl.Path += fmt.Sprintf("/%s/%s/%s", strings.Replace(groupId, ".", "/", -1), artifactId, version)

	ver := version
	if strings.HasSuffix(version, "-SNAPSHOT") {
		// get the latest snapshot
		snapshotVer, err := g.parseLastestSnapshotVersion(artifactUrl)
		if err != nil {
			return err
		}

		ver = snapshotVer
	}

	artifactUrl.Path += fmt.Sprintf("/%s-%s.%s", artifactId, ver, artType)
	dstFile := filepath.Join(filepath.Dir(dst), filepath.Base(artifactUrl.Path))

	log.Printf("Downloading %s to %s", artifactUrl, dstFile)
	return g.HttpGet.GetFile(dstFile, artifactUrl)
}

func (g *MvnGetter) parseLastestSnapshotVersion(artifactUrl *url.URL) (string, error) {
	mvnMetaUrl, err := url.Parse(artifactUrl.String())
	if err != nil {
		return "", err
	}
	mvnMetaUrl.Path += "/maven-metadata.xml"

	mvnMetaFile, err := ioutil.TempFile(os.TempDir(), "maven-metadata")
	if err != nil {
		return "", err
	}
	defer os.Remove(mvnMetaFile.Name())

	if err := g.HttpGet.GetFile(mvnMetaFile.Name(), mvnMetaUrl); err != nil {
		return "", err
	}

	mvnMetaXml, err := ioutil.ReadFile(mvnMetaFile.Name())
	if err != nil {
		return "", err
	}

	var meta Metadata
	xml.Unmarshal(mvnMetaXml, &meta)
	vers := meta.Versioning.SnapshotVersions.VersionList
	if len(vers) == 0 {
		return "", fmt.Errorf("no snapshot versions in the %s", mvnMetaUrl)
	}
	return vers[0].Value, nil
}

type Metadata struct {
	GroupId    string            `xml:"groupId"`
	ArtifactId string            `xml:"artifactId"`
	Version    string            `xml:"version"`
	Versioning SnapshotVerioning `xml:"versioning"`
}
type SnapshotVerioning struct {
	SnapshotVersions SnapshotVersions `xml:"snapshotVersions"`
}
type SnapshotVersions struct {
	VersionList []SnapshotVersion `xml:"snapshotVersion"`
}
type SnapshotVersion struct {
	Extension string `xml:"extension"`
	Value     string `xml:"value"`
}
