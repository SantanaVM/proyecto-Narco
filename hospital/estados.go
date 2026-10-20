package hospital

// Este archivo declara los estados del sistema como CONSTANTES TIPADAS.
// Usar un tipo propio (EstadoPaciente, NivelNarcolepsia, EstadoHabitacion)
// en lugar de strings sueltos hace que el compilador detecte errores:
// no se puede asignar un EstadoHabitacion donde se espera un EstadoPaciente,
// y un error de digitación en el nombre de una constante no compila.

// EstadoPaciente indica dónde y cómo está un paciente.
type EstadoPaciente int

const (
	// Despierto: el paciente está consciente.
	Despierto EstadoPaciente = iota
	// DormidoEnPasillo: le dio un ataque de sueño y aún no tiene cama.
	DormidoEnPasillo
	// DormidoEnCama: está dormido pero ya ocupa una habitación.
	DormidoEnCama
)

// String implementa fmt.Stringer para imprimir el estado de forma legible.
func (e EstadoPaciente) String() string {
	switch e {
	case Despierto:
		return "Despierto"
	case DormidoEnPasillo:
		return "Dormido en el pasillo"
	case DormidoEnCama:
		return "Dormido en cama"
	default:
		return "Estado desconocido"
	}
}

// NivelNarcolepsia indica la gravedad de la enfermedad del paciente.
type NivelNarcolepsia int

const (
	// Leve: ataques ocasionales.
	Leve NivelNarcolepsia = iota
	// Moderado: ataques frecuentes.
	Moderado
	// Severo: ataques muy frecuentes; aparece en el reporte de severidad.
	Severo
)

// String implementa fmt.Stringer.
func (n NivelNarcolepsia) String() string {
	switch n {
	case Leve:
		return "Leve"
	case Moderado:
		return "Moderado"
	case Severo:
		return "Severo"
	default:
		return "Nivel desconocido"
	}
}

// EstadoHabitacion indica si una habitación puede recibir más pacientes.
type EstadoHabitacion int

const (
	// Disponible: tiene al menos un cupo libre.
	Disponible EstadoHabitacion = iota
	// Ocupada: alcanzó su capacidad máxima.
	Ocupada
)

// String implementa fmt.Stringer.
func (e EstadoHabitacion) String() string {
	switch e {
	case Disponible:
		return "Disponible"
	case Ocupada:
		return "Ocupada"
	default:
		return "Estado desconocido"
	}
}
