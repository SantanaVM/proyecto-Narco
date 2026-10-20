package hospital

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

// ResultadoSimulacion resume lo que pasó durante la simulación concurrente.
type ResultadoSimulacion struct {
	Episodios     int            // ataques de sueño registrados
	SinHabitacion int            // veces que no hubo cama
	PorPaciente   map[string]int // id del paciente -> episodios
	Errores       []error        // errores inesperados (debería quedar vacío)
}

// resultadoPaciente es lo que cada goroutine envía por el channel al terminar.
type resultadoPaciente struct {
	id            string
	episodios     int
	sinHabitacion int
	errores       []error
}

// SimularDia lanza UNA GOROUTINE POR PACIENTE. Cada paciente, `ciclos` veces,
// espera un tiempo aleatorio, se queda dormido en una ubicación aleatoria y
// luego se despierta.
//
// Todas las goroutines modifican el mismo hospital al mismo tiempo; esto es
// seguro porque RegistrarEpisodio y DespertarPaciente toman h.mu (el Mutex)
// antes de tocar pacientes, habitaciones o historial. Si dos pacientes piden
// la última cama a la vez, el Mutex hace que uno entre primero y la ocupe;
// el segundo encuentra la habitación Ocupada y recibe ErrSinHabitacion.
//
// Patrón usado (sección 11.6 del tutorial): WaitGroup + channel. El WaitGroup
// dice CUÁNDO terminaron todas; el channel trae QUÉ produjo cada una.
func (h *Hospital) SimularDia(ciclos int, ubicaciones []string) ResultadoSimulacion {
	if len(ubicaciones) == 0 {
		ubicaciones = []string{"pasillo"}
	}
	pacientes := h.Pacientes() // copia tomada con el candado

	resultados := make(chan resultadoPaciente)
	var wg sync.WaitGroup

	for _, p := range pacientes {
		wg.Add(1) // SIEMPRE antes de lanzar la goroutine
		go func(p *Paciente) {
			defer wg.Done()
			resultados <- h.vivirPaciente(p, ciclos, ubicaciones)
		}(p)
	}

	// Goroutine coordinadora: cuando todas terminan, cierra el channel para
	// que el range de abajo finalice.
	go func() {
		wg.Wait()
		close(resultados)
	}()

	resumen := ResultadoSimulacion{PorPaciente: make(map[string]int)}
	for r := range resultados {
		resumen.Episodios += r.episodios
		resumen.SinHabitacion += r.sinHabitacion
		resumen.PorPaciente[r.id] = r.episodios
		resumen.Errores = append(resumen.Errores, r.errores...)
	}
	return resumen
}

// vivirPaciente es el cuerpo de cada goroutine. No toca ningún dato
// compartido directamente: todo pasa por métodos del Hospital con candado.
func (h *Hospital) vivirPaciente(p *Paciente, ciclos int, ubicaciones []string) resultadoPaciente {
	r := resultadoPaciente{id: p.ID()}
	for i := 0; i < ciclos; i++ {
		// Duerme la GOROUTINE (no al paciente) entre 5 y 30 ms.
		time.Sleep(time.Duration(5+rand.Intn(26)) * time.Millisecond)

		lugar := ubicaciones[rand.Intn(len(ubicaciones))]
		err := h.RegistrarEpisodio(p, lugar)
		switch {
		case err == nil:
			r.episodios++
		case errors.Is(err, ErrSinHabitacion):
			r.episodios++ // hubo episodio, solo que sin cama
			r.sinHabitacion++
		default:
			r.errores = append(r.errores, err)
			continue
		}

		time.Sleep(time.Duration(5+rand.Intn(26)) * time.Millisecond)
		if err := h.DespertarPaciente(p); err != nil {
			r.errores = append(r.errores, err)
		}
	}
	return r
}
