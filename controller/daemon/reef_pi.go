package daemon

import (
	"fmt"
	"log"
	"time"

	"github.com/reef-pi/rpi/i2c"

	"github.com/reef-pi/reef-pi/controller"
	"github.com/reef-pi/reef-pi/controller/connectors"
	"github.com/reef-pi/reef-pi/controller/drivers"
	"github.com/reef-pi/reef-pi/controller/settings"
	"github.com/reef-pi/reef-pi/controller/storage"
	"github.com/reef-pi/reef-pi/controller/telemetry"
	"github.com/reef-pi/reef-pi/controller/utils"
)

const Bucket = storage.ReefPiBucket

type ReefPi struct {
	store   storage.Store
	jacks   *connectors.Jacks
	outlets *connectors.Outlets
	inlets  *connectors.Inlets
	ais     *connectors.AnalogInputs
	drivers *drivers.Drivers

	subsystems map[string]controller.Subsystem
	settings   settings.Settings
	telemetry  telemetry.Telemetry
	version    string
	h          telemetry.HealthChecker
	bus        i2c.Bus
	a          utils.Auth
}

func New(version, database string) (*ReefPi, error) {
	store, err := storage.NewStore(database)
	if err != nil {
		log.Println("ERROR: Failed to create store. DB:", database)
		return nil, err
	}
	s, err := loadSettings(store)
	if err != nil {
		log.Println("Warning: Failed to load settings from db, Error:", err)
		log.Println("Warning: Initializing default settings in database")
		initialSettings, err := initializeSettings(store)
		if err != nil {
			return nil, err
		}
		s = initialSettings
	}
	fn := func(t, m string) error { return logError(store, t, m) }

	tele := telemetry.Initialize(Bucket, store, fn, s.Prometheus)
	bus := i2c.Bus(i2c.MockBus())
	if !s.Capabilities.DevMode {
		b, err := i2c.New()
		if err != nil {
			log.Println("ERROR: Failed to initialize i2c. Error:", err)
			logError(store, "device-i2c", "Failed to initialize i2c. Error:"+err.Error())
		} else {
			bus = b
		}
	}
	if s.RPI_PWMFreq <= 0 {
		log.Println("ERROR: Invalid  RPI PWM frequency:", s.RPI_PWMFreq, " falling back on default 100Hz")
		s.RPI_PWMFreq = 100
	}

	drvrs, err := drivers.NewDrivers(s, bus, store)
	if err != nil {
		log.Println("ERROR: failed to initialize drivers. Error:", err)
	}

	jacks := connectors.NewJacks(drvrs, store)
	outlets := connectors.NewOutlets(drvrs, store)
	inlets := connectors.NewInlets(drvrs, store)
	ais := connectors.NewAnalogInputs(drvrs, store)

	r := &ReefPi{
		bus:        bus,
		store:      store,
		settings:   s,
		telemetry:  tele,
		jacks:      jacks,
		outlets:    outlets,
		inlets:     inlets,
		ais:        ais,
		drivers:    drvrs,
		subsystems: make(map[string]controller.Subsystem),
		version:    version,
		a:          utils.NewAuth(Bucket, store),
	}
	if s.Capabilities.HealthCheck {
		r.h = telemetry.NewHealthChecker(Bucket, 1*time.Minute, s.HealthCheck, tele, store)
	}
	return r, nil
}

func (r *ReefPi) Start() error {
	if err := r.setUpErrorBucket(); err != nil {
		return err
	}
	if err := r.jacks.Setup(); err != nil {
		return err
	}
	if err := r.outlets.Setup(); err != nil {
		return err
	}
	if err := r.inlets.Setup(); err != nil {
		return err
	}
	if err := r.ais.Setup(); err != nil {
		return err
	}
	if err := r.loadSubsystems(); err != nil {
		return err
	}
	if _, err := loadDashboard(r.store); err != nil {
		initializeDashboard(r.store)
	}
	if r.settings.Capabilities.HealthCheck {
		go r.h.Start()
	}
	log.Println("reef-pi is up and running")
	return nil
}

func (r *ReefPi) unloadSubsystems() {
	for sName, sController := range r.subsystems {
		sController.Stop()
		delete(r.subsystems, sName)
		log.Println("Successfully unloaded", sName, " subsystem:")
	}
}

func (r *ReefPi) Stop() error {
	r.unloadSubsystems()
	if r.settings.Capabilities.HealthCheck {
		r.h.Stop()
	}
	r.store.Close()
	r.bus.Close()
	r.drivers.Close()
	log.Println("reef-pi is shutting down")
	return nil
}

func (r *ReefPi) Subsystem(s string) (controller.Subsystem, error) {
	sub, ok := r.subsystems[s]
	if !ok {
		return nil, fmt.Errorf("Subsystem not present: %s", s)
	}
	return sub, nil
}

func (r *ReefPi) Controller() controller.Controller {
	return controller.NewController(
		r.telemetry,
		r.store,
		r.LogError,
		r.Subsystem,
	)
}
