// Programa principal: SOLO arma el escenario de demostración (sección 6 del
// enunciado) e imprime resultados. Toda la lógica vive en el paquete hospital.
//
// Uso:
//
//	go run .                  escenario obligatorio
//	go run . -simular         escenario + simulación concurrente (bono)
//	go run -race . -simular   igual, con el detector de carreras
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"sort"

	"proyectogo/hospital"
)

func main() {
	simular := flag.Bool("simular", false, "ejecuta además la simulación concurrente")
	flag.Parse()

	escenario()

	if *simular {
		simulacion()
	}
}

// debe detiene el programa si falla la construcción del escenario (datos
// fijos escritos por nosotros: si fallan, es un error de programación).
func debe[T any](valor T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return valor
}

// personalYPacientes guarda las referencias que main necesita para imprimir.
type personalYPacientes struct {
	h         *hospital.Hospital
	medicos   []*hospital.Medico
	pacientes []*hospital.Paciente
}

// construirHospital: 2 médicos, 1 camillero, 3 habitaciones de capacidad 1
// y 5 pacientes (3 de ellos con nivel Severo).
func construirHospital() personalYPacientes {
	h := debe(hospital.NuevoHospital("Hospital de los Costeños con Narcolepsia"))

	karen := debe(hospital.NuevoMedico("M-01", "Dra. Karen Ospina", 41, "Medicina del sueño"))
	alvaro := debe(hospital.NuevoMedico("M-02", "Dr. Álvaro Mendoza", 55, "Neurología"))
	nando := debe(hospital.NuevoCamillero("C-01", "Nando Barrios (camillero)", 30))

	for _, m := range []*hospital.Medico{karen, alvaro} {
		if err := h.ContratarMedico(m); err != nil {
			log.Fatal(err)
		}
	}
	if err := h.ContratarPersonal(nando); err != nil {
		log.Fatal(err)
	}

	for _, numero := range []int{101, 102, 103} {
		if err := h.AgregarHabitacion(debe(hospital.NuevaHabitacion(numero, 1))); err != nil {
			log.Fatal(err)
		}
	}

	pacientes := []*hospital.Paciente{
		debe(hospital.NuevoPaciente("P-001", "Kelly Johana Mercado", 28, hospital.Moderado)),
		debe(hospital.NuevoPaciente("P-002", "Yeimy Padilla", 34, hospital.Severo)),
		debe(hospital.NuevoPaciente("P-003", "Jhon Jairo Pacheco", 47, hospital.Leve)),
		debe(hospital.NuevoPaciente("P-004", "Wilfrido Berrío", 62, hospital.Severo)),
		debe(hospital.NuevoPaciente("P-005", "Dayana Caraballo", 19, hospital.Severo)),
	}
	for _, p := range pacientes {
		if err := h.AdmitirPaciente(p); err != nil {
			log.Fatal(err)
		}
	}

	return personalYPacientes{
		h:         h,
		medicos:   []*hospital.Medico{karen, alvaro},
		pacientes: pacientes,
	}
}

// escenario ejecuta, en orden, la demostración obligatoria de la sección 6.
func escenario() {
	datos := construirHospital()
	h := datos.h
	p := datos.pacientes

	fmt.Printf("=== %s ===\n", h.Nombre())
	fmt.Printf("  Personal: %d | Habitaciones: %d | Pacientes: %d\n",
		len(h.Personal()), len(h.Habitaciones()), len(h.Pacientes()))

	// --- 1. Cuatro ataques de sueño en lugares distintos -------------------
	titulo("Ataques de sueño")
	ataques := []struct {
		paciente  *hospital.Paciente
		ubicacion string
	}{
		{p[0], "cafetería"},
		{p[1], "fila de radiología"},
		{p[2], "sala de espera"},
		{p[3], "pasillo 2, segundo piso"},
	}
	for _, a := range ataques {
		err := h.RegistrarEpisodio(a.paciente, a.ubicacion)
		switch {
		case err == nil:
			fmt.Printf("  %-26s se durmió en %-25s -> %s\n",
				a.paciente, a.ubicacion, a.paciente.Ubicacion())
		case errors.Is(err, hospital.ErrSinHabitacion):
			fmt.Printf("  %-26s se durmió en %-25s -> ERROR: %v\n",
				a.paciente, a.ubicacion, err)
		default:
			fmt.Printf("  error inesperado: %v\n", err)
		}
	}

	// --- Consulta 5.1 justo después de los ataques --------------------------
	titulo("Consulta 5.1: Pacientes dormidos en el pasillo")
	imprimirPasillo(h.PacientesEnPasillo())

	// --- 2. Un paciente se despierta y libera su habitación -----------------
	despierta := p[1]
	habitacion := despierta.Habitacion().Numero()
	titulo(fmt.Sprintf("%s se despierta", despierta))
	if err := h.DespertarPaciente(despierta); err != nil {
		fmt.Println("  error:", err)
	} else {
		fmt.Printf("  %s está %s y liberó la habitación %d\n",
			despierta, despierta.Estado(), habitacion)
	}

	// --- Consulta 5.3: asignar la habitación liberada -----------------------
	enPasillo := p[3]
	titulo(fmt.Sprintf("Consulta 5.3: Asignar habitación a %s", enPasillo.ID()))
	if hab, err := h.AsignarHabitacion(enPasillo); err != nil {
		fmt.Println("  ERROR:", err)
	} else {
		fmt.Printf("  %s asignado a la habitación %d (estado: %s)\n",
			enPasillo, hab.Numero(), enPasillo.Estado())
	}

	titulo("Consulta 5.1: Pacientes dormidos en el pasillo")
	imprimirPasillo(h.PacientesEnPasillo())

	// --- Consulta 5.2: historial por médico ---------------------------------
	for _, m := range datos.medicos {
		titulo("Consulta 5.2: Episodios atendidos por " + m.Nombre())
		episodios := m.MisEpisodios()
		if len(episodios) == 0 {
			fmt.Println("  (ninguno)")
		}
		for _, e := range episodios {
			fmt.Println(" ", e.Resumen())
		}
	}

	// --- Consulta 5.4: reporte de severidad ---------------------------------
	titulo("Consulta 5.4: Reporte de severidad: ")
	imprimirReporte(h.ReporteSeveridad())

	// --- Extra: el historial completo muestra el polimorfismo ---------------
	titulo("Historial completo del hospital")
	for _, e := range h.Historial() {
		fmt.Printf("  %s %s\n", e.ID(), e.Resumen())
	}
}

// simulacion ejecuta el bono concurrente sobre un hospital nuevo.
func simulacion() {
	datos := construirHospital()
	h := datos.h

	titulo("Simulación concurrente:")
	resultado := h.SimularDia(10, []string{
		"cafetería", "parqueadero", "ascensor", "baño del 2° piso", "recepción",
	})

	fmt.Printf("  Episodios registrados: %d\n", resultado.Episodios)
	fmt.Printf("  Veces sin habitación:  %d\n", resultado.SinHabitacion)
	fmt.Printf("  Episodios en historial: %d\n", len(h.Historial()))
	for _, p := range datos.pacientes {
		fmt.Printf("    %-26s %d episodios\n", p, resultado.PorPaciente[p.ID()])
	}
	for _, err := range resultado.Errores {
		fmt.Println("  error inesperado:", err)
	}
	titulo("Reporte de severidad tras la simulación")
	imprimirReporte(h.ReporteSeveridad())
}

// ---------------------------------------------------------------------------
// Funciones de impresión
// ---------------------------------------------------------------------------

func titulo(texto string) {
	fmt.Printf("\n== %s ==\n", texto)
}

func imprimirPasillo(pacientes []*hospital.Paciente) {
	if len(pacientes) == 0 {
		fmt.Println("  (ningún paciente en el pasillo)")
		return
	}
	for _, p := range pacientes {
		fmt.Printf("  %-26s (%s)\n", p, p.Ubicacion())
	}
}

// imprimirReporte ordena por id porque recorrer un mapa en Go da un orden
// distinto en cada ejecución.
func imprimirReporte(reporte map[*hospital.Paciente]int) {
	pacientes := make([]*hospital.Paciente, 0, len(reporte))
	for p := range reporte {
		pacientes = append(pacientes, p)
	}
	sort.Slice(pacientes, func(i, j int) bool {
		return pacientes[i].ID() < pacientes[j].ID()
	})
	for _, p := range pacientes {
		fmt.Printf("  %-26s %d episodio(s) hoy\n", p, reporte[p])
	}
}
