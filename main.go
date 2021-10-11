package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

func isMac() bool {
	if runtime.GOOS == "darwin" {
		return true
	}
	return false
}

func GetUidGid(username string) (int, int) {
	u, err := user.Lookup(username)
	if err != nil {
		fmt.Printf("%v", err)
		return 0, 0
	}
	gid, _ := strconv.Atoi(u.Gid)
	uid, _ := strconv.Atoi(u.Uid)
	return gid, uid
}

func runKubectl(env string, service string, num string) ([]byte, error) {
	var isMac bool = isMac()
	var cmd *exec.Cmd
	switch env := env; env {
	case "uat":
		if !isMac {
			uid, gid := GetUidGid("egame-uat")
			cmd = exec.Command("kubectl", "scale", "deploy", service, "--replicas",
				num, "-n", "egame", "--context", "uat-e-game", "--kubeconfig",
				"/home/egame-uat/.kube/config")
			cmd.SysProcAttr = &syscall.SysProcAttr{}
			cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
			fmt.Println(cmd)
		} else {
			cmd = exec.Command("kubectl", "scale", "deploy", service, "--replicas",
				num, "-n", "egame", "--context", "uat-e-game")
		}
	case "stage":
		if !isMac {
			uid, gid := GetUidGid("egame-stage")

			cmd = exec.Command("kubectl", "scale", "deploy", service, "--replicas",
				num, "-n", "egame", "--context", "egame-stage", "--kubeconfig",
				"/home/egame-stage/.kube/config")

			cmd.SysProcAttr = &syscall.SysProcAttr{}
			cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
		} else {
			cmd = exec.Command("kubectl", "scale", "deploy", service, "--replicas",
				num, "-n", "egame", "--context", "egame-cluster-stage")
		}
	case "prod":
		if !isMac {
			uid, gid := GetUidGid("egame-stage")
			cmd = exec.Command("kubectl", "scale", "deploy", service, "--replicas",
				num, "-n", "egame", "--context", "egame", "--kubeconfig",
				"/home/egame/.kube/config")
			cmd.SysProcAttr = &syscall.SysProcAttr{}
			cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
		} else {
			cmd = exec.Command("kubectl", "scale", "deploy", service, "--replicas",
				num, "-n", "egame", "--context", "egame-cluster")
		}
	}

	out, err := cmd.Output()
	//fmt.Println(string(out))
	return out, err
}

func runflux(env string, action string, service string) ([]byte, error) {
	var cmd *exec.Cmd
	var isMac bool = isMac()
	switch env := env; env {
	case "uat":
		if !isMac {
			uid, gid := GetUidGid("egame-uat")
			cmd = exec.Command("/usr/local/bin/flux", action, "hr", service, "-n", "egame",
				"--context", "uat-e-game", "--kubeconfig",
				"/home/egame-uat/.kube/config")
			cmd.SysProcAttr = &syscall.SysProcAttr{}
			cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
		} else {
			cmd = exec.Command("/usr/local/bin/flux", action, "hr", service, "-n", "egame",
				"--context", "uat-e-game")
		}
	case "stage":
		if !isMac {
			uid, gid := GetUidGid("egame-stage")
			cmd = exec.Command("/usr/local/bin/flux", action, "hr", service, "-n", "egame",
				"--context", "egame-stage", "--kubeconfig",
				"/home/egame-stage/.kube/config")

			cmd.SysProcAttr = &syscall.SysProcAttr{}
			cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}

		} else {
			cmd = exec.Command("/usr/local/bin/flux", action, "hr", service, "-n", "egame",
				"--context", "egame-cluster-stage")
		}
	case "prod":
		if !isMac {
			uid, gid := GetUidGid("egame")
			cmd = exec.Command("/usr/local/bin/flux", action, "hr", service, "-n", "egame",
				"--context", "egame-cluster", "--kubeconfig",
				"/home/egame/.kube/config")
			cmd.SysProcAttr = &syscall.SysProcAttr{}
			cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
		} else {
			cmd = exec.Command("/usr/local/bin/flux", action, "hr", service, "-n", "egame",
				"--context", "egame-cluster")
		}
	}
	out, err := cmd.Output()
	//fmt.Println(string(out))
	return out, err
}

func main() {

	fmt.Println(runtime.GOOS)
	router := gin.Default()

	router.Use(location.Default())

	// Query string parameters are parsed using the existing underlying request object.
	// The request responds to a url matching:  /welcome?firstname=Jane&lastname=Doe
	router.GET("/resume", func(c *gin.Context) {
		url := location.Get(c)
		if match, _ := regexp.MatchString("^(10|192|127|localhost)*", url.Host); match {

			env := c.Query("env")
			service := c.Query("service")
			out, err := runflux(env, "resume", service)
			//fmt.Println(err)
			c.String(http.StatusOK,
				"Action: Resume\nenvironment %s\nservice: %s\nerr: %s\n ---\n \n%s",
				env, service, err, string(out))
		}
	})
	router.POST("/resume", func(c *gin.Context) {
		token := c.PostForm("token")
		if token != "quTkKxiSXqE6qG8mdG8OuoMu" {
			c.String(http.StatusOK, "No permisson!")
			return
		}
		text := c.PostForm("text")
		env := strings.Split(text, "/")[0]
		service := strings.Split(text, "/")[1]

		c.String(http.StatusOK, "Action: Resume\nenvironment %s\nservice: %s\n",
			env, service)
		go func() {
			runflux(env, "resume", service)
		}()
	})
	router.GET("/suspend", func(c *gin.Context) {
		url := location.Get(c)
		if match, _ := regexp.MatchString("^(10|192|127|localhost)*", url.Host); match {
			env := c.Query("env")
			service := c.Query("service")
			out, err := runflux(env, "suspend", service)
			//fmt.Println(err)
			c.String(http.StatusOK,
				"Action: Suspend\nenvironment: %s\nservice: %s\nerr: %s\n ---\n \n%s",
				env, service, err, string(out))
		}
	})
	router.POST("/suspend", func(c *gin.Context) {
		token := c.PostForm("token")
		if token != "quTkKxiSXqE6qG8mdG8OuoMu" {
			c.String(http.StatusOK, "No permisson!")
			return
		}
		text := c.PostForm("text")
		env := strings.Split(text, "/")[0]
		service := strings.Split(text, "/")[1]
		c.String(http.StatusOK,
			"Action: Suspend\nenvironment %s\nservice: %s",
			env, service)
		go func() {
			runflux(env, "suspend", service)
		}()
	})
	router.GET("/scale", func(c *gin.Context) {
		url := location.Get(c)
		if match, _ := regexp.MatchString("^(10|192|127|localhost)*", url.Host); match {
			env := c.Query("env")
			service := c.Query("service")
			num := c.Query("num")
			out, err := runKubectl(env, service, num)
			//fmt.Println(err)
			c.String(http.StatusOK,
				"Action: Scale\nenvironment: %s\nservice: %s\nerr: %s ---\n \n%s",
				env, service, err, string(out))
		}
	})
	router.POST("/scale", func(c *gin.Context) {
		token := c.PostForm("token")
		if token != "quTkKxiSXqE6qG8mdG8OuoMu" {
			c.String(http.StatusOK, "No permisson!")
			return
		}
		text := c.PostForm("text")
		env := strings.Split(text, "/")[0]
		service := strings.Split(text, "/")[1]
		num := strings.Split(text, "/")[2]
		c.String(http.StatusOK,
			"Action: Scale\nenvironment: %s\nservice: %s to nums %s",
			env, service, num)
		go func() {

			runKubectl(env, service, num)

		}()
	})
	router.Run(":9090")
}
