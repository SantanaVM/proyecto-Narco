package hospital

import (
	"fmt"
	"sync/atomic"
	"time"
)

// contadorEpisodios genera ids únicos (E-001, E-002, ...). Es atómico porque
// en la simulación concurrente varias goroutines crean episodios a la vez.
var contadorEpisodios atomic.Int64

// RegistroEpisodio guarda un ataque de sueño: quién, cuándo, dónde, quién lo
// atendió y a qué habitación fue (si consiguió una).
//
// Es un VALOR inmutable (una "foto" del momento): una vez creado no tiene
// métodos que lo modifiquen, por eso se puede copiar sin riesgo al historial
// del hospital y a la lista del médico.
type RegistroEpisodio struct {
	id         string
	fechaHora  time.Time
	paciente   *Paciente
	atendedor  Atendedor   // médico, camillero o cualquier otro Atendedor
	ubicacion  string      // dónde se quedó dormido
	habitacion *Habitacion // nil si no había cama disponible
}

// nuevoRegistro arma el registro con la hora actual y la habitación que el
// paciente tenga en ese momento.
func nuevoRegistro(p *Paciente, a Atendedor, ubicacion string) RegistroEpisodio {
	n := contadorEpisodios.Add(1)
	return RegistroEpisodio{
		id:         fmt.Sprintf("E-%03d", n),
		fechaHora:  time.Now(),
		paciente:   p,
		atendedor:  a,
		ubicacion:  ubicacion,
		habitacion: p.Habitacion(),
	}
}

// Resumen retorna una línea legible, por ejemplo:
// "2026-10-19 14:15 P-001 cafetería -> habitación 101 (atendió Dra. Karen Ospina)"
func (r RegistroEpisodio) Resumen() string {
	destino := "sin habitación (pasillo)"
	if r.habitacion != nil {
		destino = fmt.Sprintf("habitación %d", r.habitacion.Numero())
	}
	return fmt.Sprintf("%s %s %-24s -> %-25s (atendió %s)",
		r.fechaHora.Format("2006-01-02 15:04"),
		r.paciente.ID(),
		r.ubicacion,
		destino,
		r.atendedor.Nombre(),
	)
}

// ID retorna el identificador del episodio.
func (r RegistroEpisodio) ID() string { return r.id }

// FechaHora retorna cuándo ocurrió el episodio.
func (r RegistroEpisodio) FechaHora() time.Time { return r.fechaHora }

// Paciente retorna el paciente que se durmió.
func (r RegistroEpisodio) Paciente() *Paciente { return r.paciente }

// Atendedor retorna quién atendió el episodio.
func (r RegistroEpisodio) Atendedor() Atendedor { return r.atendedor }

// Ubicacion retorna dónde se quedó dormido el paciente.
func (r RegistroEpisodio) Ubicacion() string { return r.ubicacion }

// Habitacion retorna la habitación asignada (nil si no hubo).
func (r RegistroEpisodio) Habitacion() *Habitacion { return r.habitacion }
