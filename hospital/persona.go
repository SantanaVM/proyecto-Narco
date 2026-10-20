package hospital

import (
	"errors"
	"strings"
)

// Persona agrupa los datos de identidad que comparten pacientes, médicos y
// camilleros. En lugar de copiar los tres campos en cada struct, Persona se
// EMBEBE (composición). Go no tiene herencia: embeber hace que los métodos
// de Persona se "promuevan" y puedan llamarse directamente, por ejemplo
// paciente.Nombre() o medico.ID().
//
// Los campos van en minúscula (no exportados): desde fuera del paquete
// hospital no se pueden leer ni modificar directamente; solo a través de
// los métodos ID(), Nombre() y Edad().
type Persona struct {
	id     string
	nombre string
	edad   int
}

// nuevaPersona valida los datos comunes. Es minúscula porque solo la usan
// los constructores de este paquete (NuevoPaciente, NuevoMedico, ...).
func nuevaPersona(id, nombre string, edad int) (Persona, error) {
	if strings.TrimSpace(id) == "" {
		return Persona{}, errors.New("el id no puede estar vacío")
	}
	if strings.TrimSpace(nombre) == "" {
		return Persona{}, errors.New("el nombre no puede estar vacío")
	}
	if edad < 0 {
		return Persona{}, errors.New("la edad no puede ser negativa")
	}
	return Persona{id: id, nombre: nombre, edad: edad}, nil
}

// Los tres métodos usan receiver por VALOR: solo consultan, no modifican.
// Un método con receiver por valor está disponible tanto para Persona como
// para *Persona, así que se promueve a Paciente, *Paciente, Medico, etc.

// ID retorna el identificador de la persona.
func (p Persona) ID() string { return p.id }

// Nombre retorna el nombre de la persona.
func (p Persona) Nombre() string { return p.nombre }

// Edad retorna la edad de la persona.
func (p Persona) Edad() int { return p.edad }
