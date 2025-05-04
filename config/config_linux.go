package config

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/apex/log"
	"github.com/cobaugh/osrelease"
	"github.com/pterodactyl/wings/system"
)

const DefaultLocation = "/etc/pterodactyl/config.yml"

// SystemConfiguration defines basic system configuration settings.
type SystemConfiguration struct {
	// The root directory where all of the pterodactyl data is stored at.
	RootDirectory string `default:"/var/lib/pterodactyl" yaml:"root_directory"`

	// Directory where logs for server installations and other wings events are logged.
	LogDirectory string `default:"/var/log/pterodactyl" yaml:"log_directory"`

	// Directory where the server data is stored at.
	Data string `default:"/var/lib/pterodactyl/volumes" yaml:"data"`

	// Directory where server archives for transferring will be stored.
	ArchiveDirectory string `default:"/var/lib/pterodactyl/archives" yaml:"archive_directory"`

	// Directory where local backups will be stored on the machine.
	BackupDirectory string `default:"/var/lib/pterodactyl/backups" yaml:"backup_directory"`

	// TmpDirectory specifies where temporary files for Pterodactyl installation processes
	// should be created. This supports environments running docker-in-docker.
	TmpDirectory string `default:"/tmp/pterodactyl" yaml:"tmp_directory"`

	// The user that should own all of the server files, and be used for containers.
	Username string `default:"pterodactyl" yaml:"username"`

	// The timezone for this Wings instance. This is detected by Wings automatically if possible,
	// and falls back to UTC if not able to be detected. If you need to set this manually, that
	// can also be done.
	//
	// This timezone value is passed into all containers created by Wings.
	Timezone string `yaml:"timezone"`

	// Definitions for the user that gets created to ensure that we can quickly access
	// this information without constantly having to do a system lookup.
	User struct {
		Uid int
		Gid int
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
		_config.System.Username = system.FirstNotEmpty(os.Getenv("WINGS_USERNAME"), "pterodactyl")
		_config.System.User.Uid = system.MustInt(system.FirstNotEmpty(os.Getenv("WINGS_UID"), "988"))
		_config.System.User.Gid = system.MustInt(system.FirstNotEmpty(os.Getenv("WINGS_GID"), "988"))
		return nil
	}

	u, err := user.Lookup(_config.System.Username)
	// If an error is returned but it isn't the unknown user error just abort
	// the process entirely. If we did find a user, return it immediately.
	if err != nil {
		if _, ok := err.(user.UnknownUserError); !ok {
			return err
		}
	} else {
		_config.System.User.Uid = system.MustInt(u.Uid)
		_config.System.User.Gid = system.MustInt(u.Gid)
		return nil
	}

	command := fmt.Sprintf("useradd --system --no-create-home --shell /usr/sbin/nologin %s", _config.System.Username)
	// Alpine Linux is the only OS we currently support that doesn't work with the useradd
	// command, so in those cases we just modify the command a bit to work as expected.
	if strings.HasPrefix(sysName, "alpine") {
		command = fmt.Sprintf("adduser -S -D -H -G %[1]s -s /sbin/nologin %[1]s", _config.System.Username)
		// We have to create the group first on Alpine, so do that here before continuing on
		// to the user creation process.
		if _, err := exec.Command("addgroup", "-S", _config.System.Username).Output(); err != nil {
			return err
		}
	}

	split := strings.Split(command, " ")
	if _, err := exec.Command(split[0], split[1:]...).Output(); err != nil {
		return err
	}
	u, err = user.Lookup(_config.System.Username)
	if err != nil {
		return err
	}
	_config.System.User.Uid = system.MustInt(u.Uid)
	_config.System.User.Gid = system.MustInt(u.Gid)
	return nil
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
		b, err := os.ReadFile("/etc/timezone")
		if err != nil {
			if !os.IsNotExist(err) {
				return errors.WithMessage(err, "config: failed to open timezone file")
			}

			_config.System.Timezone = "UTC"
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// Okay, file isn't found on this OS, we will try using timedatectl to handle this. If this
			// command fails, exit, but if it returns a value use that. If no value is returned we will
			// fall through to UTC to get Wings booted at least.
			out, err := exec.CommandContext(ctx, "timedatectl").Output()
			if err != nil {
				log.WithField("error", err).Warn("failed to execute \"timedatectl\" to determine system timezone, falling back to UTC")
				return nil
			}

			r := regexp.MustCompile(`Time zone: ([\w/]+)`)
			matches := r.FindSubmatch(out)
			if len(matches) != 2 || string(matches[1]) == "" {
				log.Warn("failed to parse timezone from \"timedatectl\" output, falling back to UTC")
				return nil
			}
			_config.System.Timezone = string(matches[1])
		} else {
			_config.System.Timezone = string(b)
		}
	}

	_config.System.Timezone = regexp.MustCompile(`(?i)[^a-z_/]+`).ReplaceAllString(_config.System.Timezone, "")
	_, err := time.LoadLocation(_config.System.Timezone)

	return errors.WithMessage(err, fmt.Sprintf("the supplied timezone %s is invalid", _config.System.Timezone))
}

// Gets the system release name.
func getSystemName() (string, error) {
	// use osrelease to get release version and ID
	release, err := osrelease.Read()
	if err != nil {
		return "", err
	}
	return release["ID"], nil
}
