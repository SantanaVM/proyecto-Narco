package hospital

// Atendedor es cualquier miembro del personal capaz de atender a un paciente
// que se quedó dormido. Es la INTERFAZ que da polimorfismo al modelo:
// el Hospital guarda su personal en un []Atendedor y despacha a quien le
// toque sin saber si es un *Medico, un *Camillero u otro tipo.
//
// En Go las interfaces se cumplen de forma implícita: basta con que un tipo
// tenga estos tres métodos; no hace falta escribir "implements".
//
// ID() y Nombre() los aporta la Persona embebida; cada tipo solo debe
// escribir su propio Atender().
type Atendedor interface {
	ID() string
	Nombre() string
	Atender(p *Paciente, ubicacion string) (RegistroEpisodio, error)
}

// Verificación en tiempo de compilación: si *Medico o *Camillero dejan de
// cumplir la interfaz, el proyecto no compila.
var (
	_ Atendedor = (*Medico)(nil)
	_ Atendedor = (*Camillero)(nil)
)
