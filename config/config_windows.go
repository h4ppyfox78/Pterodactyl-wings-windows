package config

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/pterodactyl/wings/system"
	"golang.org/x/sys/windows"
)

const DefaultLocation = "C:\\ProgramData\\Pterodactyl\\config.yml"

// SystemConfiguration defines basic system configuration settings.
type SystemConfiguration struct {
	// The root directory where all of the pterodactyl data is stored at.
	RootDirectory string `default:"C:\\ProgramData\\Pterodactyl" yaml:"root_directory"`

	// Directory where logs for server installations and other wings events are logged.
	LogDirectory string `default:"C:\\ProgramData\\Pterodactyl\\Logs" yaml:"log_directory"`

	// Directory where the server data is stored at.
	Data string `default:"C:\\ProgramData\\Pterodactyl\\Volumes" yaml:"data"`

	// Directory where server archives for transferring will be stored.
	ArchiveDirectory string `default:"C:\\ProgramData\\Pterodactyl\\Archives" yaml:"archive_directory"`

	// Directory where local backups will be stored on the machine.
	BackupDirectory string `default:"C:\\ProgramData\\Pterodactyl\\Backups" yaml:"backup_directory"`

	// TmpDirectory specifies where temporary files for Pterodactyl installation processes
	// should be created. This supports environments running docker-in-docker.
	TmpDirectory string `default:"C:\\temp\\pterodactyl" yaml:"tmp_directory"`

	// The user that should own all of the server files, and be used for containers.
	Username string `default:"Papa" yaml:"username"`

	// The timezone for this Wings instance. This is detected by Wings automatically if possible,
	// and falls back to UTC if not able to be detected. If you need to set this manually, that
	// can also be done.
	//
	// This timezone value is passed into all containers created by Wings.
	Timezone string `yaml:"timezone"`

	// Definitions for the user that gets created to ensure that we can quickly access
	// this information without constantly having to do a system lookup.
	User struct {
		Uid string
		Gid string
	}

	// The amount of time in seconds that can elapse before a server's disk space calculation is
	// considered stale and a re-check should occur. DANGER: setting this value too low can seriously
	// impact system performance and cause massive I/O bottlenecks and high CPU usage for the Wings
	// process.
	//
	// Set to 0 to disable disk checking entirely. This will always return 0 for the disk space used
	// by a server and should only be set in extreme scenarios where performance is critical and
	// disk usage is not a concern.
	DiskCheckInterval int64 `default:"150" yaml:"disk_check_interval"`

	// If set to true, file permissions for a server will be checked when the process is
	// booted. This can cause boot delays if the server has a large amount of files. In most
	// cases disabling this should not have any major impact unless external processes are
	// frequently modifying a servers' files.
	CheckPermissionsOnBoot bool `default:"true" yaml:"check_permissions_on_boot"`

	// If set to false Wings will not attempt to write a log rotate configuration to the disk
	// when it boots and one is not detected.
	EnableLogRotate bool `default:"true" yaml:"enable_log_rotate"`

	// The number of lines to send when a server connects to the websocket.
	WebsocketLogCount int `default:"150" yaml:"websocket_log_count"`

	Sftp SftpConfiguration `yaml:"sftp"`

	CrashDetection CrashDetection `yaml:"crash_detection"`

	Backups Backups `yaml:"backups"`

	Transfers Transfers `yaml:"transfers"`
}

// EnsurePterodactylUser ensures that the Pterodactyl core user exists on the
// system. This user will be the owner of all data in the root data directory
// and is used as the user within containers. If files are not owned by this
// user there will be issues with permissions on Docker mount points.
//
// This function IS NOT thread safe and should only be called in the main thread
// when the application is booting.
func EnsurePterodactylUser() error {
	sysName, err := getSystemName()
	if err != nil {
		return err
	}

	// Our way of detecting if wings is running inside of Docker.
	if sysName == "distroless" {
		_config.System.Username = system.FirstNotEmpty(os.Getenv("WINGS_USERNAME"), "Papa")
		_config.System.User.Uid = system.FirstNotEmpty(os.Getenv("WINGS_UID"), "988")
		_config.System.User.Gid = system.FirstNotEmpty(os.Getenv("WINGS_GID"), "988")
		return nil
	}

	u, err := user.Lookup(_config.System.Username)

	// If an error is returned but it isn't the unknown user error just abort
	// the process entirely. If we did find a user, return it immediately.
	// golang.org.x/sys/windows.ERROR_NONE_MAPPED (1332)
	if err == nil {
		_config.System.Username = strings.Split(u.Username, "\\")[1]
		_config.System.User.Uid = u.Uid
		_config.System.User.Gid = u.Gid
		return nil
	} else if err != windows.ERROR_NONE_MAPPED {
		return err
	}

	command := fmt.Sprintf("net user %s /add", _config.System.Username)

	split := strings.Split(command, " ")
	if _, err := exec.Command(split[0], split[1:]...).Output(); err != nil {
		return err
	}

	if u, err := user.Lookup(_config.System.Username); err != nil {
		return err
	} else {
		_config.System.Username = strings.Split(u.Username, "\\")[1]
		_config.System.User.Uid = u.Uid
		_config.System.User.Gid = u.Gid
		return nil
	}
}

// ConfigureTimezone sets the timezone data for the configuration if it is
// currently missing. If a value has been set, this functionality will only run
// to validate that the timezone being used is valid.
//
// This function IS NOT thread-safe.
func ConfigureTimezone() error {
	tz := os.Getenv("TZ")
	if _config.System.Timezone == "" && tz != "" {
		_config.System.Timezone = tz
	}
	if _config.System.Timezone == "" {
		_config.System.Timezone = time.Now().Location().String()
	}

	_config.System.Timezone = regexp.MustCompile(`(?i)[^a-z_/]+`).ReplaceAllString(_config.System.Timezone, "")
	_, err := time.LoadLocation(_config.System.Timezone)

	return errors.WithMessage(err, fmt.Sprintf("the supplied timezone %s is invalid", _config.System.Timezone))
}

// Gets the system release name.
func getSystemName() (string, error) {
	//TODO Find way to get correct information on Windows
	return "", nil
}
