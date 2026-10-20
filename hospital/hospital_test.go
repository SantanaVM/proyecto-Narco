// Las pruebas usan el paquete "hospital_test" (externo): solo ven lo
// exportado, igual que main. Así también se comprueba el encapsulamiento.
package hospital_test

import (
	"errors"
	"fmt"
	"testing"

	"proyectogo/hospital"
)

// ---------------------------------------------------------------------------
// Ayudantes
// ---------------------------------------------------------------------------

// nuevoHospital arma un hospital con 1 médico, 1 camillero y las
// habitaciones indicadas (todas de capacidad 1).
func nuevoHospital(t *testing.T, habitaciones ...int) (*hospital.Hospital, *hospital.Medico) {
	t.Helper()
	h, err := hospital.NuevoHospital("Hospital de prueba")
	if err != nil {
		t.Fatal(err)
	}
	m, err := hospital.NuevoMedico("M-01", "Dra. Prueba", 40, "Sueño")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.ContratarMedico(m); err != nil {
		t.Fatal(err)
	}
	c, err := hospital.NuevoCamillero("C-01", "Camillero Prueba", 30)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.ContratarPersonal(c); err != nil {
		t.Fatal(err)
	}
	for _, n := range habitaciones {
		hab, err := hospital.NuevaHabitacion(n, 1)
		if err != nil {
			t.Fatal(err)
		}
		if err := h.AgregarHabitacion(hab); err != nil {
			t.Fatal(err)
		}
	}
	return h, m
}

func admitir(t *testing.T, h *hospital.Hospital, id string, nivel hospital.NivelNarcolepsia) *hospital.Paciente {
	t.Helper()
	p, err := hospital.NuevoPaciente(id, "Paciente "+id, 30, nivel)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.AdmitirPaciente(p); err != nil {
		t.Fatal(err)
	}
	return p
}

// ---------------------------------------------------------------------------
// Pruebas obligatorias
// ---------------------------------------------------------------------------

// 1) Asignación de habitación cuando hay una libre.
func TestAsignarHabitacionConCamaLibre(t *testing.T) {
	h, _ := nuevoHospital(t, 101)
	p := admitir(t, h, "P-1", hospital.Leve)

	if err := p.SufrirAtaqueDeSueno("cafetería"); err != nil {
		t.Fatal(err)
	}
	hab, err := h.AsignarHabitacion(p)
	if err != nil {
		t.Fatalf("se esperaba habitación, llegó error: %v", err)
	}
	if hab.Numero() != 101 {
		t.Errorf("habitación = %d, se esperaba 101", hab.Numero())
	}
	if hab.Estado() != hospital.Ocupada {
		t.Errorf("estado habitación = %s, se esperaba Ocupada", hab.Estado())
	}
	if p.Estado() != hospital.DormidoEnCama {
		t.Errorf("estado paciente = %s, se esperaba DormidoEnCama", p.Estado())
	}
}

// 2) Falla cuando no hay habitación: error, sin pánico, paciente en pasillo.
func TestAsignarHabitacionSinCamaLibre(t *testing.T) {
	h, _ := nuevoHospital(t, 101)
	p1 := admitir(t, h, "P-1", hospital.Leve)
	p2 := admitir(t, h, "P-2", hospital.Severo)

	if err := h.RegistrarEpisodio(p1, "cafetería"); err != nil {
		t.Fatal(err)
	}
	err := h.RegistrarEpisodio(p2, "pasillo 2")
	if !errors.Is(err, hospital.ErrSinHabitacion) {
		t.Fatalf("error = %v, se esperaba ErrSinHabitacion", err)
	}
	if p2.Estado() != hospital.DormidoEnPasillo {
		t.Errorf("estado = %s, se esperaba DormidoEnPasillo", p2.Estado())
	}
	if p2.Habitacion() != nil {
		t.Errorf("se inventó la habitación %d", p2.Habitacion().Numero())
	}
	if n := len(h.Habitaciones()); n != 1 {
		t.Errorf("hay %d habitaciones, debe seguir habiendo 1", n)
	}
	// El episodio sí queda registrado aunque no haya cama.
	if n := len(h.Historial()); n != 2 {
		t.Errorf("historial tiene %d episodios, se esperaban 2", n)
	}
}

// 3) Consulta 5.1: solo aparecen los que están dormidos en el pasillo.
func TestPacientesEnPasillo(t *testing.T) {
	h, _ := nuevoHospital(t, 101)
	p1 := admitir(t, h, "P-1", hospital.Leve)
	p2 := admitir(t, h, "P-2", hospital.Leve)
	admitir(t, h, "P-3", hospital.Leve) // se queda despierto

	_ = h.RegistrarEpisodio(p1, "cafetería") // obtiene la 101
	_ = h.RegistrarEpisodio(p2, "ascensor")  // queda en el pasillo

	enPasillo := h.PacientesEnPasillo()
	if len(enPasillo) != 1 || enPasillo[0] != p2 {
		t.Fatalf("en pasillo = %v, se esperaba solo P-2", enPasillo)
	}
}

// ---------------------------------------------------------------------------
// Pruebas adicionales
// ---------------------------------------------------------------------------

// Consulta 5.2: el médico conoce sus episodios sin recibir el hospital.
func TestMisEpisodios(t *testing.T) {
	h, m := nuevoHospital(t, 101, 102, 103)
	p1 := admitir(t, h, "P-1", hospital.Leve)
	p2 := admitir(t, h, "P-2", hospital.Leve)
	p3 := admitir(t, h, "P-3", hospital.Leve)

	// Round-robin: médico, camillero, médico.
	for _, p := range []*hospital.Paciente{p1, p2, p3} {
		if err := h.RegistrarEpisodio(p, "cafetería"); err != nil {
			t.Fatal(err)
		}
	}
	episodios := m.MisEpisodios()
	if len(episodios) != 2 {
		t.Fatalf("el médico tiene %d episodios, se esperaban 2", len(episodios))
	}
	if episodios[0].Paciente() != p1 || episodios[1].Paciente() != p3 {
		t.Errorf("pacientes inesperados en los episodios del médico")
	}
	// El camillero no diagnostica: P-2 no debe tener médico.
	if p2.MedicoAsignado() != nil {
		t.Errorf("el camillero no debería asignar médico")
	}
	if p1.MedicoAsignado() != m {
		t.Errorf("P-1 debería quedar a cargo del médico")
	}
}

// Consulta 5.4: cruza pacientes severos con el conteo de episodios.
func TestReporteSeveridad(t *testing.T) {
	h, _ := nuevoHospital(t, 101)
	severo := admitir(t, h, "P-1", hospital.Severo)
	sinEpisodios := admitir(t, h, "P-2", hospital.Severo)
	leve := admitir(t, h, "P-3", hospital.Leve)

	for i := 0; i < 3; i++ {
		if err := h.RegistrarEpisodio(severo, "cafetería"); err != nil {
			t.Fatal(err)
		}
		if err := h.DespertarPaciente(severo); err != nil {
			t.Fatal(err)
		}
	}
	_ = h.RegistrarEpisodio(leve, "parqueadero")

	reporte := h.ReporteSeveridad()
	if len(reporte) != 2 {
		t.Fatalf("el reporte tiene %d pacientes, se esperaban 2 (solo severos)", len(reporte))
	}
	if reporte[severo] != 3 {
		t.Errorf("P-1 = %d episodios, se esperaban 3", reporte[severo])
	}
	if n, esta := reporte[sinEpisodios]; !esta || n != 0 {
		t.Errorf("P-2 debe aparecer con 0 episodios (está=%v, n=%d)", esta, n)
	}
	if _, esta := reporte[leve]; esta {
		t.Errorf("un paciente Leve no debe aparecer en el reporte")
	}
}

// Despertar libera la habitación y permite reasignarla (escenario sección 6).
func TestDespertarLiberaHabitacion(t *testing.T) {
	h, _ := nuevoHospital(t, 101)
	p1 := admitir(t, h, "P-1", hospital.Leve)
	p2 := admitir(t, h, "P-2", hospital.Leve)

	_ = h.RegistrarEpisodio(p1, "cafetería")
	_ = h.RegistrarEpisodio(p2, "pasillo") // sin cama

	if err := h.DespertarPaciente(p1); err != nil {
		t.Fatal(err)
	}
	if p1.Habitacion() != nil || p1.Estado() != hospital.Despierto {
		t.Fatalf("P-1 debería estar despierto y sin habitación")
	}
	hab, err := h.AsignarHabitacion(p2)
	if err != nil {
		t.Fatalf("la habitación liberada debía asignarse: %v", err)
	}
	if hab.Numero() != 101 {
		t.Errorf("habitación = %d, se esperaba 101", hab.Numero())
	}
	if err := h.DespertarPaciente(p1); err == nil {
		t.Errorf("despertar a alguien despierto debe retornar error")
	}
}

// Un tipo definido SOLO en la prueba también puede ser contratado: el
// hospital no conoce los tipos concretos, solo la interfaz Atendedor.
type enfermeraDePrueba struct {
	id, nombre string
	atendidos  int
}

func (e *enfermeraDePrueba) ID() string     { return e.id }
func (e *enfermeraDePrueba) Nombre() string { return e.nombre }
func (e *enfermeraDePrueba) Atender(p *hospital.Paciente, _ string) (hospital.RegistroEpisodio, error) {
	e.atendidos++
	return hospital.RegistroEpisodio{}, errors.New("la enfermera de prueba nunca puede atender")
}

func TestDespachoPolimorfico(t *testing.T) {
	h, err := hospital.NuevoHospital("Solo enfermera")
	if err != nil {
		t.Fatal(err)
	}
	enf := &enfermeraDePrueba{id: "E-1", nombre: "Enfermera"}
	if err := h.ContratarPersonal(enf); err != nil {
		t.Fatal(err)
	}
	p := admitir(t, h, "P-1", hospital.Leve)

	err = h.RegistrarEpisodio(p, "cafetería")
	if !errors.Is(err, hospital.ErrSinPersonal) {
		t.Fatalf("error = %v, se esperaba ErrSinPersonal", err)
	}
	if enf.atendidos != 1 {
		t.Errorf("el hospital debió despachar a la enfermera una vez")
	}
}

func TestPacienteNoAdmitido(t *testing.T) {
	h, _ := nuevoHospital(t, 101)
	extraño, _ := hospital.NuevoPaciente("X-1", "Extraño", 20, hospital.Leve)
	if _, err := h.AsignarHabitacion(extraño); !errors.Is(err, hospital.ErrPacienteNoAdmitido) {
		t.Errorf("error = %v, se esperaba ErrPacienteNoAdmitido", err)
	}
}

// Prueba de tabla para los String() de las constantes tipadas.
func TestEstadosString(t *testing.T) {
	casos := []struct {
		valor  fmt.Stringer
		quiere string
	}{
		{hospital.Despierto, "Despierto"},
		{hospital.DormidoEnPasillo, "Dormido en el pasillo"},
		{hospital.DormidoEnCama, "Dormido en cama"},
		{hospital.Severo, "Severo"},
		{hospital.Ocupada, "Ocupada"},
	}
	for _, c := range casos {
		if got := c.valor.String(); got != c.quiere {
			t.Errorf("String() = %q, se esperaba %q", got, c.quiere)
		}
	}
}

// Simulación concurrente: ejecutar con  go test -race ./...
func TestSimulacionConcurrenteSinCarreras(t *testing.T) {
	h, _ := nuevoHospital(t, 101, 102)
	for _, id := range []string{"P-1", "P-2", "P-3", "P-4", "P-5"} {
		admitir(t, h, id, hospital.Severo)
	}

	r := h.SimularDia(5, []string{"cafetería", "parqueadero"})

	if len(r.Errores) > 0 {
		t.Fatalf("errores inesperados: %v", r.Errores)
	}
	if r.Episodios != 25 {
		t.Errorf("episodios = %d, se esperaban 25 (5 pacientes × 5 ciclos)", r.Episodios)
	}
	if n := len(h.Historial()); n != r.Episodios {
		t.Errorf("historial = %d, no coincide con %d episodios", n, r.Episodios)
	}
	// Al final todos despertaron: ninguna habitación puede quedar ocupada.
	for _, hab := range h.Habitaciones() {
		if !hab.EstaDisponible() {
			t.Errorf("la habitación %d quedó ocupada", hab.Numero())
		}
	}
}
