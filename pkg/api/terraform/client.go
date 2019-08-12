package terraform

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform/config/module"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-null/null"
)

// Platform is the platform to be managed by Terraform
type Platform struct {
	Code         string
	Providers    map[string]terraform.ResourceProvider
	Provisioners map[string]terraform.ResourceProvisioner
	Vars         map[string]interface{}
	State        *State
}

// State is an alias for terraform.State
type State = terraform.State

// NewPlatform return an instance of Platform with default values
func NewPlatform(code string) *Platform {
	platform := &Platform{
		Code: code,
	}
	platform.Providers = defaultProviders()
	platform.Provisioners = defaultProvisioners()

	platform.State = terraform.NewState()

	return platform
}

func defaultProviders() map[string]terraform.ResourceProvider {
	return map[string]terraform.ResourceProvider{
		"null": null.Provider(),
	}
}

// AddProvider adds a new provider to the providers list
func (p *Platform) AddProvider(name string, provider terraform.ResourceProvider) *Platform {
	p.Providers[name] = provider
	return p
}

func defaultProvisioners() map[string]terraform.ResourceProvisioner {
	return map[string]terraform.ResourceProvisioner{}
}

// AddProvisioner adds a new provisioner to the provisioner list
func (p *Platform) AddProvisioner(name string, provisioner terraform.ResourceProvisioner) *Platform {
	p.Provisioners[name] = provisioner
	return p
}

// BindVars binds the map of variables to the Platform variables, to be used
// by Terraform
func (p *Platform) BindVars(vars map[string]interface{}) *Platform {
	for name, value := range vars {
		p.Var(name, value)
	}

	return p
}

// Var set a variable with it's value
func (p *Platform) Var(name string, value interface{}) *Platform {
	if len(p.Vars) == 0 {
		p.Vars = make(map[string]interface{})
	}
	p.Vars[name] = value

	return p
}

// WriteState takes a io.Writer as input to write the Terraform state
func (p *Platform) WriteState(w io.Writer) (*Platform, error) {
	return p, terraform.WriteState(p.State, w)
}

// ReadState takes a io.Reader as input to read from it the Terraform state
func (p *Platform) ReadState(r io.Reader) (*Platform, error) {
	state, err := terraform.ReadState(r)
	if err != nil {
		return p, err
	}
	p.State = state
	return p, nil
}

// WriteStateToFile save the state of the Terraform state to a file
func (p *Platform) WriteStateToFile(filename string) (*Platform, error) {
	var state bytes.Buffer
	if _, err := p.WriteState(&state); err != nil {
		return p, err
	}
	return p, ioutil.WriteFile(filename, state.Bytes(), 0644)
}

// ReadStateFromFile will load the Terraform state from a file and assign it to the
// Platform state.
func (p *Platform) ReadStateFromFile(filename string) (*Platform, error) {
	file, err := os.Open(filename)
	if err != nil {
		return p, err
	}
	return p.ReadState(file)
}

// Apply brings the platform to the desired state. It'll destroy the platform
// when `destroy` is `true`.
func (p *Platform) Apply(destroy bool) error {
	ctx, err := p.newContext(destroy)
	if err != nil {
		return err
	}

	// state := ctx.State()

	if _, err := ctx.Refresh(); err != nil {
		return err
	}

	if _, err := ctx.Plan(); err != nil {
		return err
	}

	_, err = ctx.Apply()
	p.State = ctx.State()

	return err
}

// Plan returns execution plan for an existing configuration to apply to the
// platform.
func (p *Platform) Plan(destroy bool) (*terraform.Plan, error) {
	ctx, err := p.newContext(destroy)
	if err != nil {
		return nil, err
	}

	if _, err := ctx.Refresh(); err != nil {
		return nil, err
	}

	plan, err := ctx.Plan()
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// newContext creates the Terraform context or configuration
func (p *Platform) newContext(destroy bool) (*terraform.Context, error) {
	module, err := p.module()
	if err != nil {
		return nil, err
	}

	providerResolver := p.getProviderResolver()
	provisioners := p.getProvisioners()

	// Create ContextOpts with the current state and variables to apply
	ctxOpts := &terraform.ContextOpts{
		Destroy:          destroy,
		State:            p.State,
		Variables:        p.Vars,
		Module:           module,
		ProviderResolver: providerResolver,
		Provisioners:     provisioners,
	}

	ctx, err := terraform.NewContext(ctxOpts)
	if err != nil {
		return nil, err
	}

	// TODO: Validate the context

	return ctx, nil
}

func (p *Platform) module() (*module.Tree, error) {
	if len(p.Code) == 0 {
		return nil, fmt.Errorf("no code to apply")
	}

	// Get a temporal directory to save the infrastructure code
	cfgPath, err := ioutil.TempDir("", ".terranova")
	if err != nil {
		return nil, err
	}
	// This defer is executed second
	defer os.RemoveAll(cfgPath)

	// Save the infrastructure code
	cfgFileName := filepath.Join(cfgPath, "main.tf")
	cfgFile, err := os.Create(cfgFileName)
	if err != nil {
		return nil, err
	}
	// This defer is executed first
	defer cfgFile.Close()
	if _, err = io.Copy(cfgFile, strings.NewReader(p.Code)); err != nil {
		return nil, err
	}

	mod, err := module.NewTreeModule("", cfgPath)
	if err != nil {
		return nil, err
	}

	s := module.NewStorage(filepath.Join(cfgPath, "modules"), nil)
	s.Mode = module.GetModeNone // or module.GetModeGet?

	if err := mod.Load(s); err != nil {
		return nil, fmt.Errorf("failed to load the modules. %s", err)
	}

	if err := mod.Validate().Err(); err != nil {
		return nil, fmt.Errorf("failed Terraform code validation. %s", err)
	}

	return mod, nil
}

func (p *Platform) getProviderResolver() terraform.ResourceProviderResolver {
	ctxProviders := make(map[string]terraform.ResourceProviderFactory)

	for name, provider := range p.Providers {
		ctxProviders[name] = terraform.ResourceProviderFactoryFixed(provider)
	}

	providerResolver := terraform.ResourceProviderResolverFixed(ctxProviders)

	// TODO: Reset the providers?

	return providerResolver
}

func (p *Platform) getProvisioners() map[string]terraform.ResourceProvisionerFactory {
	provisioners := make(map[string]terraform.ResourceProvisionerFactory)

	for name, provisioner := range p.Provisioners {
		provisioners[name] = func() (terraform.ResourceProvisioner, error) {
			return provisioner, nil
		}
	}

	return provisioners
}
