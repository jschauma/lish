/*
 * Copyright (c) 2014, Twitter. Inc
 *
 * Originally written by Jan Schaumann <jschauma@netmeister.org> in October
 * 2014.
 *
 * A very simple, limited shell.  Allowed commands are read from a
 * configuration file.
 */

package main

/* remember to set CC=clang on OS X */

/*
#include <unistd.h>
*/
import "C"

import (
	"bufio"
	"fmt"
	"io"
	"log/syslog"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
)

const EXIT_FAILURE = 1
const EXIT_SUCCESS = 0
const FORBIDDEN = 127

const PATH = "/bin:/usr/bin:/sbin:/usr/sbin:/usr/local/bin:/usr/local/sbin"

const CONFIG = "/etc/lishrc"
const CONFIG_DIR = "/etc/lish"
const PROGNAME = "lish"
const VERSION = "0.2"

var ALLOWED_COMMANDS = map[string]bool{}
var GIVEN_COMMANDS string
var LAST_RVAL int
var LOGGER *syslog.Writer
var USER string
var TRACE bool = false
var TTY bool = false

/*
 * Main
 */

func main() {
	var err error
	TTY = isatty(os.Stdin.Fd())

	getopts()
	getUser()
	parseConfigs()

	if ssh_cmd := os.Getenv("SSH_ORIGINAL_COMMAND"); len(ssh_cmd) > 0 {
		GIVEN_COMMANDS = ssh_cmd
	}

	setEnvironment()

	LOGGER, err = syslog.New(syslog.LOG_INFO|syslog.LOG_AUTH, PROGNAME)
	if err != nil {
		fail(fmt.Sprintf("Unable to open syslog: %v\n", err))
	}

	if len(GIVEN_COMMANDS) > 0 {
		processInput(GIVEN_COMMANDS)
	} else {
		readExecLoop()
	}
	os.Exit(LAST_RVAL)
	/* NOTREACHED */
}

/*
 * Functions
 */

func argcheck(flag string, args []string, i int) {
	if len(args) <= (i + 1) {
		fail(fmt.Sprintf("'%v' needs an argument\n", flag))
	}
}

func builtin(cmd []string) (handled bool) {
	handled = false

	if TRACE {
		fmt.Fprintf(os.Stderr, "+ %s\n", strings.Join(cmd, " "))
	}

	if cmd[0] == "cd" {
		doCd(cmd)
		return true
	}

	if cmd[0] == "exit" {
		os.Exit(EXIT_SUCCESS)
		/* NOTREACHED */
	}

	return
}

func doCd(cmd []string) {
	var dir string
	if len(cmd) > 2 {
		fail("Too many arguments to builtin 'cd'.\n")
	} else if len(cmd) == 2 {
		dir = cmd[1]
	} else {
		u, err := user.Current()
		if err != nil {
			fail(fmt.Sprintf("Unable to get current user: %s\n", err))
		}
		dir = u.HomeDir
	}

	if err := os.Chdir(dir); err != nil {
		fail(fmt.Sprintf("Unable to change directory to '%s': %s\n", dir, err))
	}
	return
}

func fail(msg string) {
	fmt.Fprintf(os.Stderr, msg)
	os.Exit(EXIT_FAILURE)
	/* NOTREACHED */
}

func getopts() {
	eatit := false
	args := os.Args[1:]
	for i, arg := range args {
		if eatit {
			eatit = false
			continue
		}
		switch arg {
		case "-V":
			printVersion()
			os.Exit(EXIT_SUCCESS)
			/* NOTREACHED */
		case "-c":
			eatit = true
			argcheck("-c", args, i)
			GIVEN_COMMANDS = args[i+1]
		case "-x":
			TRACE = true
		default:
			fmt.Fprintf(os.Stderr, "Unexpected option or argument: %v\n", args[i])
			usage(os.Stderr)
			os.Exit(EXIT_FAILURE)
			/* NOTREACHED */
		}
	}
}

func getUser() {
	user, err := user.Current()
	if err != nil {
		fail(fmt.Sprintf("Unable to get current user: %v\n", err))
	}
	USER = user.Username
}

func isatty(fd uintptr) bool {
	return int(C.isatty(C.int(fd))) != 0
}

func parseConfigs() {
	userfile := CONFIG_DIR + "/" + USER
	configfiles := []string{CONFIG, userfile}
	for _, file := range configfiles {
		parseFile(file)
	}
}

func parseFile(file string) {
	if TRACE {
		fmt.Fprintf(os.Stderr, "++ %s\n", file)
	}

	fd, err := os.Open(file)
	if err != nil {
		return
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), "#")[0]
		cmd := strings.Fields(line)
		if len(cmd) > 0 {
			ALLOWED_COMMANDS[strings.Join(cmd, " ")] = true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
	}
}

func printVersion() {
	fmt.Printf("%v version %v\n", PROGNAME, VERSION)
}

func prompt() {
	fmt.Printf("%s$ ", PROGNAME)
}

func processCommand(given string) (retval int) {
	retval = EXIT_SUCCESS
	given_cmd := strings.Fields(given)
	/* squeeze extra spaces */
	given = strings.Join(given_cmd, " ")

	if len(given) < 1 {
		return
	}

	if builtin(given_cmd) {
		return
	}

	ok := false

	if ALLOWED_COMMANDS[given] {
		ok = true
	} else {
		for cmd, _ := range ALLOWED_COMMANDS {
			if strings.TrimSpace(cmd) == strings.TrimSpace(given) {
				ok = true
				break
			}
			if strings.Contains(cmd, "*") {
				cmd = strings.TrimSpace(strings.Split(cmd, "*")[0])
				if strings.HasPrefix(given, cmd) {
					ok = true
					break
				}
			}
		}
	}

	if !ok {
		retval = FORBIDDEN
		LOGGER.Err(fmt.Sprintf("%s not allowed to run '%s'", USER, given))
		fmt.Fprintf(os.Stderr, "Not allowed: '%s'\n", given)
		return
	}

	return runCommand(given_cmd)
}

func processInput(input string) {
	for _, cmd := range strings.Split(input, ";") {
		if len(cmd) > 0 {
			LAST_RVAL = processCommand(cmd)
		}
	}
}

func readExecLoop() {

	LAST_RVAL = EXIT_SUCCESS
	if TTY {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGQUIT,
			syscall.SIGSTOP)
		go func() {
			for _ = range sigc {
				fmt.Printf("\n")
				prompt()
			}
		}()
		prompt()
	}

	for {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := scanner.Text()
			processInput(input)
			if TTY {
				prompt()
			}
		}

		err := scanner.Err()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan error on stdin: %v\n", err)
		} else {
			if TTY {
				fmt.Printf("\n")
			}
			break
		}
	}

	return
}

func runCommand(cmd []string) (retval int) {
	retval = EXIT_SUCCESS

	cwd := "unknown"
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get current working directory: %v\n", err)
	}

	command := strings.Join(cmd, " ")
	LOGGER.Info(fmt.Sprintf("%s: %s: '%s'", USER, cwd, command))
	if TRACE {
		fmt.Fprintf(os.Stderr, "+ %s\n", command)
	}

	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			retval = exitError.Sys().(syscall.WaitStatus).ExitStatus()
		}
	}

	return
}

func setEnvironment() {
	os.Clearenv()

	if err := os.Setenv("PATH", PATH); err != nil {
		fail(fmt.Sprintf("Unable to set PATH: %v\n", err))
	}

	if err := os.Setenv("USER", USER); err != nil {
		fail(fmt.Sprintf("Unable to set USER: %v\n", err))
	}

	if err := os.Setenv("SHELL", os.Args[0]); err != nil {
		fail(fmt.Sprintf("Unable to set SHELL: %v\n", err))
	}
}

func usage(out io.Writer) {
	usage := `Usage: %v [-Vx] [-c command]
	-V          print version information and exit
	-c command  execute the given command
	-x          enable tracing
`
	fmt.Fprintf(out, usage, PROGNAME)
}
