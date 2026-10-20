package hospital

import (
	"errors"
	"fmt"
	"strings"
)

// Paciente es un costeño con narcolepsia internado en el hospital.
type Paciente struct {
	Persona // embebida por valor: Paciente reutiliza ID(), Nombre() y Edad()

	nivel      NivelNarcolepsia
	estado     EstadoPaciente
	ubicacion  string      // dónde está ahora (o dónde se quedó dormido)
	medico     *Medico     // médico a cargo; nil si aún no tiene
	habitacion *Habitacion // habitación que ocupa; nil si no tiene cama
}

// NuevoPaciente crea un paciente despierto, sin médico y sin habitación.
func NuevoPaciente(id, nombre string, edad int, nivel NivelNarcolepsia) (*Paciente, error) {
	persona, err := nuevaPersona(id, nombre, edad)
	if err != nil {
		return nil, fmt.Errorf("paciente inválido: %w", err)
	}
	return &Paciente{
		Persona:   persona,
		nivel:     nivel,
		estado:    Despierto,
		ubicacion: "recepción",
	}, nil
}

// SufrirAtaqueDeSueno registra que el paciente se quedó dormido de golpe en
// la ubicación indicada. Queda en el pasillo hasta que le asignen cama.
// Receiver por puntero: modifica el paciente original, no una copia.
func (p *Paciente) SufrirAtaqueDeSueno(ubicacion string) error {
	if strings.TrimSpace(ubicacion) == "" {
		return errors.New("la ubicación del ataque no puede estar vacía")
	}
	if p.estado != Despierto {
		return fmt.Errorf("%s ya está dormido (%s)", p.id, p.estado)
	}
	p.estado = DormidoEnPasillo
	p.ubicacion = ubicacion
	return nil
}

// Despertar pone al paciente despierto. Si estaba en una cama, libera la
// habitación para que otro paciente pueda usarla.
func (p *Paciente) Despertar() error {
	if p.estado == Despierto {
		return fmt.Errorf("%s ya está despierto", p.id)
	}
	if p.habitacion != nil {
		if err := p.habitacion.Liberar(p); err != nil {
			return err
		}
		p.habitacion = nil
	}
	p.estado = Despierto
	return nil
}

// acostarEn pasa al paciente a la cama de la habitación dada. Es minúscula:
// solo el Hospital (mismo paquete) puede decidir a qué cama va un paciente.
func (p *Paciente) acostarEn(h *Habitacion) {
	p.habitacion = h
	p.estado = DormidoEnCama
	p.ubicacion = fmt.Sprintf("habitación %d", h.Numero())
}

// Estado retorna el estado actual del paciente.
func (p *Paciente) Estado() EstadoPaciente { return p.estado }

// Nivel retorna el nivel de narcolepsia.
func (p *Paciente) Nivel() NivelNarcolepsia { return p.nivel }

// Ubicacion retorna dónde está el paciente.
func (p *Paciente) Ubicacion() string { return p.ubicacion }

// MedicoAsignado retorna el médico a cargo (puede ser nil).
func (p *Paciente) MedicoAsignado() *Medico { return p.medico }

// Habitacion retorna la habitación que ocupa (puede ser nil).
func (p *Paciente) Habitacion() *Habitacion { return p.habitacion }

// String implementa fmt.Stringer: "P-001 Yeimy Padilla".
func (p *Paciente) String() string {
	return fmt.Sprintf("%s %s", p.id, p.nombre)
}
