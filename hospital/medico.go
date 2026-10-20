package hospital

import (
	"errors"
	"fmt"
	"strings"
)

// Medico diagnostica pacientes y atiende emergencias de sueño.
// La relación médico -> episodios se guarda DENTRO del médico, por eso
// MisEpisodios() no necesita recibir el hospital como parámetro.
type Medico struct {
	Persona // embebida: aporta ID(), Nombre() y Edad()

	especialidad string
	pacientes    []*Paciente        // pacientes a su cargo
	episodios    []RegistroEpisodio // episodios que atendió
}

// NuevoMedico crea un médico sin pacientes ni episodios.
func NuevoMedico(id, nombre string, edad int, especialidad string) (*Medico, error) {
	persona, err := nuevaPersona(id, nombre, edad)
	if err != nil {
		return nil, fmt.Errorf("médico inválido: %w", err)
	}
	if strings.TrimSpace(especialidad) == "" {
		return nil, errors.New("médico inválido: la especialidad no puede estar vacía")
	}
	return &Medico{Persona: persona, especialidad: especialidad}, nil
}

// DiagnosticarPaciente deja al paciente bajo el cuidado de este médico.
// Si ya era su paciente no hace nada; si pertenece a otro médico, falla.
func (d *Medico) DiagnosticarPaciente(p *Paciente) error {
	if p == nil {
		return errors.New("no se puede diagnosticar un paciente nil")
	}
	if p.medico == d {
		return nil
	}
	if p.medico != nil {
		return fmt.Errorf("%s ya está a cargo de %s", p.ID(), p.medico.Nombre())
	}
	p.medico = d
	d.pacientes = append(d.pacientes, p)
	return nil
}

// AtenderEmergenciaDeSueno verifica que el paciente realmente esté dormido y,
// si todavía no tiene médico, lo diagnostica.
func (d *Medico) AtenderEmergenciaDeSueno(p *Paciente) error {
	if p == nil {
		return errors.New("no se puede atender un paciente nil")
	}
	if p.Estado() == Despierto {
		return fmt.Errorf("%s está despierto: no hay emergencia que atender", p.ID())
	}
	if p.MedicoAsignado() == nil {
		return d.DiagnosticarPaciente(p)
	}
	return nil
}

// Atender cumple la interfaz Atendedor: atiende la emergencia y guarda el
// episodio en la lista propia del médico.
func (d *Medico) Atender(p *Paciente, ubicacion string) (RegistroEpisodio, error) {
	if err := d.AtenderEmergenciaDeSueno(p); err != nil {
		return RegistroEpisodio{}, err
	}
	registro := nuevoRegistro(p, d, ubicacion)
	d.episodios = append(d.episodios, registro)
	return registro, nil
}

// MisEpisodios retorna una copia de todos los episodios en que intervino.
func (d *Medico) MisEpisodios() []RegistroEpisodio {
	copia := make([]RegistroEpisodio, len(d.episodios))
	copy(copia, d.episodios)
	return copia
}

// Especialidad retorna la especialidad del médico.
func (d *Medico) Especialidad() string { return d.especialidad }

// Pacientes retorna una copia de los pacientes a su cargo.
func (d *Medico) Pacientes() []*Paciente {
	copia := make([]*Paciente, len(d.pacientes))
	copy(copia, d.pacientes)
	return copia
}
