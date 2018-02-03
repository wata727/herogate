package herogate

import (
	"io/ioutil"
	"os"
	"testing"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
)

func TestDetectAppFromRepo(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get current directory: " + err.Error())
	}
	defer os.Chdir(currentDir)

	dir, err := ioutil.TempDir("", "validRepository")
	if err != nil {
		t.Fatal("Failed to create tempdir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal("Failed to init git reporisoty: " + err.Error())
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "herogate",
		URLs: []string{"ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/testApp"},
	})
	if err != nil {
		t.Fatal("Failed to create remote: " + err.Error())
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal("Failed to change directory: " + err.Error())
	}

	region, app := detectAppFromRepo()
	if region != "us-east-1" {
		t.Fatalf("Expected region is `us-east-1`, but get `%s`", region)
	}
	if app != "testApp" {
		t.Fatalf("Expected app is `testApp`, but get `%s`", app)
	}
}

func TestDetectAppFromRepo__nonRepository(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get current directory: " + err.Error())
	}
	defer os.Chdir(currentDir)

	dir, err := ioutil.TempDir("", "nonRepository")
	if err != nil {
		t.Fatal("Failed to create tempdir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal("Failed to change directory: " + err.Error())
	}

	region, app := detectAppFromRepo()
	if region != "" {
		t.Fatalf("Expected region is empty, but get `%s`", region)
	}
	if app != "" {
		t.Fatalf("Expected app is empty, but get `%s`", app)
	}
}

func TestDetectAppFromRepo__noRemotes(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get current directory: " + err.Error())
	}
	defer os.Chdir(currentDir)

	dir, err := ioutil.TempDir("", "noRemotes")
	if err != nil {
		t.Fatal("Failed to create tempdir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	_, err = git.PlainInit(dir, false)
	if err != nil {
		t.Fatal("Failed to init git reporisoty: " + err.Error())
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal("Failed to change directory: " + err.Error())
	}

	region, app := detectAppFromRepo()
	if region != "" {
		t.Fatalf("Expected region is empty, but get `%s`", region)
	}
	if app != "" {
		t.Fatalf("Expected app is empty, but get `%s`", app)
	}
}

func TestDetectAppFromRepo__invalidRemoteURL(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get current directory: " + err.Error())
	}
	defer os.Chdir(currentDir)

	dir, err := ioutil.TempDir("", "invalidRemoteURL")
	if err != nil {
		t.Fatal("Failed to create tempdir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal("Failed to init git reporisoty: " + err.Error())
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "herogate",
		URLs: []string{"git@github.com:us-east-1/testApp"},
	})
	if err != nil {
		t.Fatal("Failed to create remote: " + err.Error())
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal("Failed to change directory: " + err.Error())
	}

	region, app := detectAppFromRepo()
	if region != "" {
		t.Fatalf("Expected region is empty, but get `%s`", region)
	}
	if app != "" {
		t.Fatalf("Expected app is empty, but get `%s`", app)
	}
}
