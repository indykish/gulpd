package machine

import (
	"encoding/json"
	"net"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	nsqp "github.com/crackcomm/nsqueue/producer"
	"github.com/megamsys/gulp/carton"
	"github.com/megamsys/gulp/db"
	"github.com/megamsys/gulp/meta"
	"github.com/megamsys/gulp/provision"
)

const (
	TOPIC          = "vms"
	SSHFILESBUCKET = "sshfiles"
)

type Machine struct {
	Name      string
	Id        string
	CartonId  string
	CartonsId string
	Level     provision.BoxLevel
	SSH       provision.BoxSSH
	PublicIp  string
	Status    provision.Status
}

func (m *Machine) SetStatus(status provision.Status) error {
	log.Debugf("  set status[%s] of machine (%s, %s)", m.Id, m.Name, status.String())

	if asm, err := carton.NewAmbly(m.CartonId); err != nil {
		return err
	} else if err = asm.SetStatus(status); err != nil {

		return err
	}

	if m.Level == provision.BoxSome {
		log.Debugf("  set status[%s] of machine (%s, %s)", m.Id, m.Name, status.String())

		if comp, err := carton.NewComponent(m.Id); err != nil {
			return err
		} else if err = comp.SetStatus(status); err != nil {
			return err
		}
	}
	return nil
}

// FindAndSetIps returns the non loopback local IP4 (can be public or private)
// we also have to add it in for ipv6
func (m *Machine) FindAndSetIps() error {
	ips := m.findIps()

	log.Debugf("  find and setips of machine (%s, %s)", m.Id, m.Name)

	if asm, err := carton.NewAmbly(m.CartonId); err != nil {
		return err
	} else if err = asm.NukeAndSetOutputs(ips); err != nil {
		return err
	}
	return nil
}

// FindIps returns the non loopback local IP4 (can be public or private)
// if an iface contains a string "pub", then we consider it a public interface
func (m *Machine) findIps() map[string][]string {
	var ips = make(map[string][]string)
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	pubipv4s := []string{}
	priipv4s := []string{}
	for _, iface := range ifaces {
		ifaddress, err := iface.Addrs()
		if err != nil {
			return ips
		}
		for _, address := range ifaddress {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					if strings.Contains(iface.Name, "eth0") {
						pubipv4s = append(pubipv4s, ipnet.IP.String())
					} else {
						priipv4s = append(priipv4s, ipnet.IP.String())
					}
				}
			}
		}
	}
	ips[carton.PUBLICIPV4] = pubipv4s
	ips[carton.PRIVATEIPV4] = priipv4s
	return ips
}

// append user sshkey into authorized_keys file
func (m *Machine) AppendAuthKeys() error {
	sshkey, err := db.FetchObject(SSHFILESBUCKET, m.SSH.Pub())
	if err != nil {
		return err
	}

	f, err := os.OpenFile(m.SSH.AuthKeysFile(), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(sshkey); err != nil {
		return err
	}
	return nil
}

func (m *Machine) ChangeState(status provision.Status) error {
	log.Debugf("  change state of machine (%s, %s)", m.Name, status.String())

	pons := nsqp.New()
	if err := pons.Connect(meta.MC.NSQd[0]); err != nil {
		return err
	}

	bytes, err := json.Marshal(
		carton.Requests{
			CatId:     m.CartonsId,
			Action:    status.String(),
			Category:  carton.STATE,
			CreatedAt: time.Now().Local().Format(time.RFC822),
		})

	if err != nil {
		return err
	}

	log.Debugf("  pub to topic (%s, %s)", TOPIC, bytes)
	if err = pons.Publish(TOPIC, bytes); err != nil {
		return err
	}

	defer pons.Stop()
	return nil
}
