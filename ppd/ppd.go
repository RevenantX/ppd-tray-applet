package ppd

import (
	"errors"
	"fmt"

	"github.com/godbus/dbus/v5"
)

type Client struct {
	conn  *dbus.Conn
	dest  string
	path  dbus.ObjectPath
	iface string
}

func Connect() (*Client, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("connect system bus: %w", err)
	}

	c, err := detect(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return c, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func (c *Client) ServiceName() string         { return c.dest }
func (c *Client) ObjectPath() dbus.ObjectPath { return c.path }
func (c *Client) Interface() string           { return c.iface }

func (c *Client) ActiveProfile() (string, error) {
	v, err := c.obj().GetProperty(c.iface + ".ActiveProfile")
	if err != nil {
		return "", err
	}
	s, ok := v.Value().(string)
	if !ok {
		return "", fmt.Errorf("ActiveProfile has unexpected type %T", v.Value())
	}
	return s, nil
}

func (c *Client) Profiles() ([]map[string]dbus.Variant, error) {
	v, err := c.obj().GetProperty(c.iface + ".Profiles")
	if err != nil {
		return nil, err
	}
	arr, ok := v.Value().([]map[string]dbus.Variant)
	if !ok {
		return nil, fmt.Errorf("Profiles has unexpected type %T", v.Value())
	}
	return arr, nil
}

func (c *Client) SetActiveProfile(profile string) error {
	return c.obj().SetProperty(c.iface+".ActiveProfile", dbus.MakeVariant(profile))
}

func (c *Client) SubscribeActiveProfileChanges() (<-chan string, func(), error) {
	if c.conn == nil {
		return nil, nil, errors.New("dbus connection not initialized")
	}

	signalCh := make(chan *dbus.Signal, 10)
	c.conn.Signal(signalCh)

	addErr := c.conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
		dbus.WithMatchObjectPath(c.path),
	)
	if addErr != nil {
		c.conn.RemoveSignal(signalCh)
		return nil, nil, addErr
	}

	out := make(chan string, 5)
	done := make(chan struct{})

	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case sig := <-signalCh:
				if sig == nil || len(sig.Body) < 3 {
					continue
				}

				ifaceName, ok := sig.Body[0].(string)
				if !ok || ifaceName != c.iface {
					continue
				}

				changed, ok := sig.Body[1].(map[string]dbus.Variant)
				if !ok {
					continue
				}

				if v, ok := changed["ActiveProfile"]; ok {
					if s, ok := v.Value().(string); ok {
						out <- s
					}
				}
			}
		}
	}()

	stop := func() {
		close(done)
		c.conn.RemoveSignal(signalCh)
		_ = c.conn.RemoveMatchSignal(
			dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
			dbus.WithMatchMember("PropertiesChanged"),
			dbus.WithMatchObjectPath(c.path),
		)
	}

	return out, stop, nil
}

// --- internals ---

func (c *Client) obj() dbus.BusObject {
	return c.conn.Object(c.dest, c.path)
}

func detect(conn *dbus.Conn) (*Client, error) {
	// Prefer the newer namespace if present, fallback to older.
	candidates := []struct {
		dest  string
		path  dbus.ObjectPath
		iface string
	}{
		{"org.freedesktop.UPower.PowerProfiles", "/org/freedesktop/UPower/PowerProfiles", "org.freedesktop.UPower.PowerProfiles"},
		{"net.hadess.PowerProfiles", "/net/hadess/PowerProfiles", "net.hadess.PowerProfiles"},
	}

	for _, cand := range candidates {
		//nameHasOwner
		var ok bool
		err := conn.BusObject().Call("org.freedesktop.DBus.NameHasOwner", 0, cand.dest).Store(&ok)
		if err != nil {
			return nil, err
		}
		if ok {
			return &Client{
				conn:  conn,
				dest:  cand.dest,
				path:  cand.path,
				iface: cand.iface,
			}, nil
		}
	}

	return nil, errors.New("no PowerProfiles service found on system bus")
}
