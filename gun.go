package mgun

import (
	"time"
	"regexp"
	"fmt"
	"math/rand"
	"strings"
)

var (
	arrayParamRegexp = regexp.MustCompile(`[\w\d\-\_]\[\]+`)
	configParamRegexp = regexp.MustCompile(`\$\{([\w\d\-\_\.]+)\}`)
	gun = &Gun{
		Features: make(Features, 0),
		Calibers: make(CaliberMap),
		Cartridges: make(Cartridges, 0),
	}
)

type Gun struct {
	Features     Features      `yaml:"headers"`
	Calibers     CaliberMap    `yaml:"params"`
	Cartridges   Cartridges    `yaml:"requests"`
}

func GetGun() *Gun {
	return gun
}

func (this *Gun) prepare() {
	if len(this.Cartridges) == 0 {
		cartridge := new(Cartridge)
		cartridge.path = NewNamedDescribedFeature(GET_METHOD, "/")
		this.Cartridges = append(this.Cartridges, cartridge)
	}
	reporter.log("cartridges count - %v", this.Cartridges)
}

func (this *Gun) findCaliber(path string) *Caliber {
	parts := strings.Split(path, ".")
	if caliber, ok := this.Calibers[parts[0]]; ok {
		return this.findInCaliber(caliber, parts[1:])
	} else {
		return nil
	}
}

func (this *Gun) findInCaliber(caliber *Caliber, pathParts []string) *Caliber {
	var nextPathParts []string
	if len(pathParts) > 1 {
		nextPathParts = pathParts[1:]
	} else {
		nextPathParts = pathParts
	}
	switch caliber.kind {
	case CALIBER_KIND_LIST:
		calibers := caliber.feature.description.(CaliberList)
		rand.Seed(time.Now().UnixNano())
		randCaliber := calibers[rand.Intn(len(calibers)) + 0];
		if randCaliber.kind == CALIBER_KIND_MAP {
			return this.findInCaliber(randCaliber, nextPathParts)
		} else {
			return randCaliber
		}
	case CALIBER_KIND_MAP:
		caliberMap := caliber.feature.description.(CaliberMap)
		if childCaliber, ok := caliberMap[pathParts[0]]; ok {
			return this.findInCaliber(childCaliber, nextPathParts)
		} else {
			return nil
		}
	case CALIBER_KIND_SESSION:
		return nil
	case CALIBER_KIND_SIMPLES:
		return nil
	default:
		return caliber
	}
}

type CaliberList []*Caliber
type CaliberMap map[string]*Caliber
type CaliberMapList []CaliberMap

func (this CaliberMap) UnmarshalYAML(unmarshal func(yaml interface{}) error) error {
	reporter.log("unmarchal calibers")

	calibers := make(map[interface{}]interface{})
	err := unmarshal(calibers)

	if rawSessions, ok := calibers["session"].([]interface{}); ok {
		reporter.log("fill session calibers")

		delete(calibers, "session")
		list := make(CaliberMapList, 0)

		for _, rawSession := range rawSessions {
			if session, ok := rawSession.(map[interface{}]interface{}); ok {
				reporter.ln()
				caliberMap := make(CaliberMap)
				caliberMap.fill(session)
				list = append(list, caliberMap)
			}
		}

		caliber := new(Caliber)
		caliber.kind = CALIBER_KIND_SESSION
		caliber.children = list
		this["session"] = caliber
	}

	reporter.ln()
	reporter.log("fill other calibers")
	this.fill(calibers)
	return err
}

func (this CaliberMap) fill(rawCalibers map[interface{}]interface{}) {
	for rawKey, rawValue := range rawCalibers {
		var caliber *Caliber
		key := fmt.Sprintf("%v", rawKey)

		switch rawValue.(type) {
		case []interface{}:
			if arrayParamRegexp.MatchString(key) {
				caliber = NewCaliberByKindAndFeature(CALIBER_KIND_SIMPLES, NewDescribedFeature(rawValue))
			} else {
				list := make(CaliberList, 0)
				list.fill(rawValue.([]interface{}))
				caliber = NewCaliberByKindAndFeature(CALIBER_KIND_LIST, NewDescribedFeature(list))
			}
			break
		case map[interface{}]interface{}:
			caliberMap := make(CaliberMap)
			caliberMap.fill(rawValue.(map[interface{}]interface{}))
			caliber = NewCaliberByKindAndFeature(CALIBER_KIND_MAP, NewDescribedFeature(caliberMap))
			break
		default:
			caliber = NewCaliberByKindAndFeature(CALIBER_KIND_SIMPLE, NewDescribedFeature(rawValue))
			break
		}
		this[key] = caliber
		reporter.log("caliber: key - %v,  kind - %v, feature - %v, children - %v", key, caliber.kind, caliber.feature, caliber.children)
	}
}

func (this *CaliberList) fill(rawList []interface{}) {
	for _, value := range rawList {
		switch value.(type) {
		case []interface{}:
			list := make(CaliberList, 0)
			list.fill(value.([]interface{}))
			*this = append(*this, NewCaliberByKindAndFeature(CALIBER_KIND_LIST, NewDescribedFeature(list)))
			break
		case map[interface{}]interface{}:
			caliberMap := make(CaliberMap)
			caliberMap.fill(value.(map[interface{}]interface{}))
			*this = append(*this, NewCaliberByKindAndFeature(CALIBER_KIND_MAP, NewDescribedFeature(caliberMap)))
			break
		default:
			*this = append(*this, NewCaliberByKindAndFeature(CALIBER_KIND_SIMPLE, NewDescribedFeature(value)))
			break
		}
	}
}

type Caliber struct {
	kind     CaliberKind
	feature  *Feature
	children interface{}
}

func NewCaliber() *Caliber {
	return new(Caliber)
}

func NewCaliberByKind(kind CaliberKind) *Caliber {
	caliber := NewCaliber()
	caliber.kind = kind
	return caliber
}

func NewCaliberByKindAndFeature(kind CaliberKind, feature *Feature) *Caliber {
	caliber := NewCaliberByKind(kind)
	caliber.feature = feature
	return caliber
}

type CaliberKind int

const (
	CALIBER_KIND_SIMPLE CaliberKind = iota
	CALIBER_KIND_SIMPLES
	CALIBER_KIND_MAP
	CALIBER_KIND_LIST
	CALIBER_KIND_SESSION
)

type Cartridges []*Cartridge

func (this *Cartridges) UnmarshalYAML(unmarshal func(yaml interface{}) error) error {
	reporter.ln()
	reporter.log("unmarchal cartridges")

	rawCartridges := make([]interface{}, 0)
	err := unmarshal(&rawCartridges)

	this.fill(rawCartridges)

	return err
}

func (this *Cartridges) fill(rawCartridges []interface{}) {
	for _, rawCartridge := range rawCartridges {
		cartridge := new(Cartridge)
		for rawKey, rawValue := range rawCartridge.(map[interface{}]interface{}) {
			key := rawKey.(string)
			switch (key) {
			case GET_METHOD, POST_METHOD, PUT_METHOD, DELETE_METHOD:
				cartridge.path = NewNamedDescribedFeature(key, rawValue)
				kill.shotsCount++
				reporter.ln()
				reporter.log("shots count: %v", kill.shotsCount)
				reporter.ln()
				break;
			case RANDOM_METHOD, SYNC_METHOD:
				cartridge.path = NewNamedFeature(key)
				cartridge.children = make(Cartridges, 0)
				cartridge.children.fill(rawValue.([]interface{}))
				break;
			case "headers":
				cartridge.bulletFeatures = make(Features, 0)
				cartridge.bulletFeatures.fill(rawValue.(map[interface{}]interface{}))
				break;
			case "params":
				cartridge.chargeFeatures = make(Features, 0)
				cartridge.chargeFeatures.fill(rawValue.(map[interface{}]interface{}))
				break;
			case "timeout":
				cartridge.timeout = time.Duration(rawValue.(int))
				break;
			}
		}
		*this = append(*this, cartridge)
		reporter.log(
			"cartridge: path - %v,  bulletFeatures - %v, chargeFeatures - %v, timeout - %v, children - %v",
			cartridge.path,
			cartridge.bulletFeatures,
			cartridge.chargeFeatures,
			cartridge.timeout,
			cartridge.children,
		)
	}
}

const (
	GET_METHOD     = "GET"
	POST_METHOD    = "POST"
	PUT_METHOD     = "PUT"
	DELETE_METHOD  = "DELETE"
	RANDOM_METHOD  = "RANDOM"
 	SYNC_METHOD    = "SYNC"
	INCLUDE_METHOD = "INCLUDE"
)

type Cartridge struct {
	path           *Feature
	bulletFeatures Features
	chargeFeatures Features
	timeout        time.Duration
	children       Cartridges
}

func (this *Cartridge) GetMethod() string {
	return this.path.name
}

func (this *Cartridge) GetPathAsString() string {
	return this.path.String()
}

func (this *Cartridge) GetChildren() Cartridges {
	if this.path.name == RANDOM_METHOD {
		for i := range this.children {
			j := rand.Intn(i + 1)
			this.children[i], this.children[j] = this.children[j], this.children[i]
		}
		return this.children
	} else if this.path.name == SYNC_METHOD {
		return this.children
	} else {
		return Cartridges{}
	}
}

type FeatureKind int

const (
	FEATURE_KIND_SIMPLE FeatureKind = iota
	FEATURE_KIND_MULTIPLE
)

type Features []*Feature

func (this *Features) UnmarshalYAML(unmarshal func(yaml interface{}) error) error {
	rawFeatures := make(map[interface{}]interface{})
	err := unmarshal(&rawFeatures)

	this.fill(rawFeatures)

	return err
}

func (this *Features) fill(rawFeatures map[interface{}]interface{}) {
	for rawKey, rawValue := range rawFeatures {
		key := rawKey.(string)
		value := rawValue.(string)
		*this = append(*this, NewNamedDescribedFeature(key, value))
	}
}

type Feature struct {
	name        string
	description interface{}
	units       []string
	kind        FeatureKind
}

func NewFeature() *Feature {
	return new(Feature)
}

func NewNamedFeature(name string) *Feature {
	feature := NewFeature()
	feature.name = name
	return feature
}

func NewDescribedFeature(rawDescription interface{}) *Feature {
	return NewFeature().setDescription(rawDescription)
}

func NewNamedDescribedFeature(name string, rawDescription interface{}) *Feature {
	return NewNamedFeature(name).setDescription(rawDescription)
}

func (this *Feature) setDescription(rawDescription interface{}) *Feature {
	switch rawDescription.(type) {
	case string:
		description := rawDescription.(string)
		if configParamRegexp.MatchString(description) {
			this.description = configParamRegexp.ReplaceAllString(description, "%v")
			this.units = make([]string, 0)
			for _, submatches := range configParamRegexp.FindAllStringSubmatch(description, -1) {
				if len(submatches) >= 2 {
					this.units = append(this.units, submatches[1])
				}
			}
			this.kind = FEATURE_KIND_MULTIPLE
		} else {
			this.setSimpleDescription(description)
		}
		break
	default:
		this.setSimpleDescription(rawDescription)
		break
	}
	return this
}

func (this *Feature) setSimpleDescription(description interface{}) *Feature {
	this.description = description
	this.kind = FEATURE_KIND_SIMPLE
	return this
}

func (this *Feature) String() string {
	if this.kind == FEATURE_KIND_SIMPLE {
		return fmt.Sprintf("%v", this.description)
	} else {
		values := make([]interface{}, len(this.units))
		for i, unit := range this.units {
			caliber := gun.findCaliber(unit)
			if caliber != nil && caliber.feature != nil {
				values[i] = caliber.feature.String()
			}
		}
		return fmt.Sprintf(this.description.(string), values...)
	}
}
