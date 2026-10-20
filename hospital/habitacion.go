package hospital

import (
	"errors"
	"fmt"
)

// Habitacion es un cuarto con una o más camas.
type Habitacion struct {
	numero    int
	capacidad int
	estado    EstadoHabitacion
	ocupantes []*Paciente
}

// NuevaHabitacion crea una habitación vacía y disponible.
func NuevaHabitacion(numero, capacidad int) (*Habitacion, error) {
	if numero <= 0 {
		return nil, errors.New("el número de habitación debe ser positivo")
	}
	if capacidad <= 0 {
		return nil, errors.New("la capacidad debe ser al menos 1")
	}
	return &Habitacion{
		numero:    numero,
		capacidad: capacidad,
		estado:    Disponible,
		// make con capacidad: el slice nunca necesitará crecer.
		ocupantes: make([]*Paciente, 0, capacidad),
	}, nil
}

// EstaDisponible indica si queda al menos una cama libre.
func (h *Habitacion) EstaDisponible() bool {
	return h.estado == Disponible
}

// Ocupar mete al paciente en la habitación. Falla si ya está llena o si el
// paciente ya estaba adentro.
func (h *Habitacion) Ocupar(p *Paciente) error {
	if p == nil {
		return errors.New("no se puede ocupar la habitación con un paciente nil")
	}
	if !h.EstaDisponible() {
		return fmt.Errorf("la habitación %d está ocupada", h.numero)
	}
	for _, o := range h.ocupantes {
		if o == p {
			return fmt.Errorf("%s ya está en la habitación %d", p.ID(), h.numero)
		}
	}
	h.ocupantes = append(h.ocupantes, p)
	h.actualizarEstado()
	return nil
}

// Liberar saca al paciente de la habitación y la deja disponible.
func (h *Habitacion) Liberar(p *Paciente) error {
	for i, o := range h.ocupantes {
		if o == p {
			// Quita el elemento i uniendo lo que hay antes y después de él.
			h.ocupantes = append(h.ocupantes[:i], h.ocupantes[i+1:]...)
			h.actualizarEstado()
			return nil
		}
	}
	return fmt.Errorf("%s no está en la habitación %d", p.ID(), h.numero)
}

// actualizarEstado mantiene el estado coherente con el número de ocupantes.
func (h *Habitacion) actualizarEstado() {
	if len(h.ocupantes) >= h.capacidad {
		h.estado = Ocupada
	} else {
		h.estado = Disponible
	}
}

// Numero retorna el número de la habitación.
func (h *Habitacion) Numero() int { return h.numero }

// Capacidad retorna cuántos pacientes caben.
func (h *Habitacion) Capacidad() int { return h.capacidad }

// Estado retorna Disponible u Ocupada.
func (h *Habitacion) Estado() EstadoHabitacion { return h.estado }

// Ocupantes retorna una COPIA del slice, para que nadie desde fuera pueda
// modificar la lista interna de la habitación.
func (h *Habitacion) Ocupantes() []*Paciente {
	copia := make([]*Paciente, len(h.ocupantes))
	copy(copia, h.ocupantes)
	return copia
}
