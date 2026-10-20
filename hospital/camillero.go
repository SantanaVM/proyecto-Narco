package hospital

import (
	"errors"
	"fmt"
)

// Camillero traslada pacientes dormidos, pero NO puede diagnosticar.
// Es el segundo tipo concreto que cumple la interfaz Atendedor.
type Camillero struct {
	Persona // embebida: aporta ID(), Nombre() y Edad()

	traslados int // cuántos pacientes ha atendido
}

// NuevoCamillero crea un camillero.
func NuevoCamillero(id, nombre string, edad int) (*Camillero, error) {
	persona, err := nuevaPersona(id, nombre, edad)
	if err != nil {
		return nil, fmt.Errorf("camillero inválido: %w", err)
	}
	return &Camillero{Persona: persona}, nil
}

// Atender cumple la interfaz Atendedor. A diferencia del médico, no
// diagnostica ni le asigna médico al paciente: solo lo recoge y registra.
func (c *Camillero) Atender(p *Paciente, ubicacion string) (RegistroEpisodio, error) {
	if p == nil {
		return RegistroEpisodio{}, errors.New("no se puede atender un paciente nil")
	}
	if p.Estado() == Despierto {
		return RegistroEpisodio{}, fmt.Errorf("%s está despierto: no hay a quién cargar", p.ID())
	}
	c.traslados++
	return nuevoRegistro(p, c, ubicacion), nil
}

// Traslados retorna cuántos pacientes ha atendido el camillero.
func (c *Camillero) Traslados() int { return c.traslados }
