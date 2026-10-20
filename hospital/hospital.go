// Package hospital contiene el modelo del "Hospital de los Costeños con
// Narcolepsia": pacientes, médicos, camilleros, habitaciones y episodios.
//
// El paquete no imprime nada: sus métodos retornan datos y errores, y es el
// programa principal (main) quien decide cómo mostrarlos.
package hospital

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Errores "centinela": main y las pruebas pueden reconocerlos con errors.Is.
var (
	// ErrSinHabitacion se retorna cuando todas las habitaciones están llenas.
	ErrSinHabitacion = errors.New("no hay habitación disponible")
	// ErrPacienteNoAdmitido se retorna al operar sobre un paciente ajeno.
	ErrPacienteNoAdmitido = errors.New("el paciente no está admitido en este hospital")
	// ErrSinPersonal se retorna si nadie puede atender un episodio.
	ErrSinPersonal = errors.New("no hay personal disponible para atender")
)

// Hospital es la entidad central: conoce a todo el personal, a los
// pacientes, las habitaciones y el historial completo de episodios.
//
// mu (un sync.Mutex) protege todo el estado del hospital. Todos los métodos
// exportados de Hospital toman el candado, así que es seguro llamarlos desde
// varias goroutines a la vez (ver simulacion.go).
type Hospital struct {
	mu sync.Mutex

	nombre       string
	medicos      []*Medico
	personal     []Atendedor // médicos, camilleros... (polimorfismo)
	pacientes    []*Paciente // slice: conserva el orden de admisión
	porID        map[string]*Paciente
	habitaciones []*Habitacion
	historial    []RegistroEpisodio
	turno        int // índice del próximo Atendedor a despachar (round-robin)
}

// NuevoHospital crea un hospital vacío.
func NuevoHospital(nombre string) (*Hospital, error) {
	if strings.TrimSpace(nombre) == "" {
		return nil, errors.New("el hospital necesita un nombre")
	}
	return &Hospital{
		nombre: nombre,
		porID:  make(map[string]*Paciente),
	}, nil
}

// ---------------------------------------------------------------------------
// Registro de personal, habitaciones y pacientes
// ---------------------------------------------------------------------------

// ContratarMedico agrega un médico al hospital y también al personal que
// atiende emergencias.
func (h *Hospital) ContratarMedico(d *Medico) error {
	if d == nil {
		return errors.New("no se puede contratar un médico inexistente")
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.agregarPersonal(d); err != nil {
		return err
	}
	h.medicos = append(h.medicos, d)
	return nil
}

// ContratarPersonal agrega cualquier Atendedor (camillero, enfermera, ...).
// Recibe la interfaz, no un tipo concreto: un tipo nuevo que cumpla
// Atendedor se puede contratar sin cambiar este método.
func (h *Hospital) ContratarPersonal(a Atendedor) error {
	if a == nil {
		return errors.New("no se puede contratar personal nil")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.agregarPersonal(a)
}

// agregarPersonal asume que el candado ya está tomado.
func (h *Hospital) agregarPersonal(a Atendedor) error {
	for _, existente := range h.personal {
		if existente.ID() == a.ID() {
			return fmt.Errorf("ya existe personal con id %s", a.ID())
		}
	}
	h.personal = append(h.personal, a)
	return nil
}

// AgregarHabitacion registra una habitación nueva.
func (h *Hospital) AgregarHabitacion(hab *Habitacion) error {
	if hab == nil {
		return errors.New("no se puede agregar una habitación nil")
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, existente := range h.habitaciones {
		if existente.Numero() == hab.Numero() {
			return fmt.Errorf("la habitación %d ya existe", hab.Numero())
		}
	}
	h.habitaciones = append(h.habitaciones, hab)
	return nil
}

// AdmitirPaciente registra un paciente. Usa el mapa porID para rechazar ids
// repetidos en tiempo constante.
func (h *Hospital) AdmitirPaciente(p *Paciente) error {
	if p == nil {
		return errors.New("no se puede admitir un paciente nil")
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, existe := h.porID[p.ID()]; existe {
		return fmt.Errorf("ya existe un paciente con id %s", p.ID())
	}
	h.porID[p.ID()] = p
	h.pacientes = append(h.pacientes, p)
	return nil
}

// ---------------------------------------------------------------------------
// Operaciones del día a día
// ---------------------------------------------------------------------------

// RegistrarEpisodio es el flujo completo de un ataque de sueño:
//
//  1. el paciente se duerme en la ubicación (queda en el pasillo),
//  2. se intenta asignarle la primera habitación libre,
//  3. se despacha al siguiente miembro del personal (sin importar su tipo),
//  4. el episodio queda en el historial del hospital.
//
// Si no hay cama, el episodio SÍ se registra (el paciente fue atendido en el
// pasillo) y el método retorna un error que envuelve a ErrSinHabitacion.
func (h *Hospital) RegistrarEpisodio(p *Paciente, ubicacion string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.verificarAdmitido(p); err != nil {
		return err
	}
	if len(h.personal) == 0 {
		return ErrSinPersonal
	}

	// 1. El ataque de sueño.
	if err := p.SufrirAtaqueDeSueno(ubicacion); err != nil {
		return err
	}

	// 2. Buscar cama. Guardamos el error para reportarlo al final, pero el
	//    paciente igual debe ser atendido en el pasillo.
	_, errHabitacion := h.asignarHabitacion(p)

	// 3. Despachar personal de forma polimórfica.
	registro, err := h.despachar(p, ubicacion)
	if err != nil {
		return err
	}

	// 4. Guardar en el historial general.
	h.historial = append(h.historial, registro)
	return errHabitacion
}

// despachar recorre el personal empezando por h.turno (round-robin) y usa al
// primero que logre atender. Aquí está el POLIMORFISMO: a.Atender(...) llama
// al método del tipo concreto (Medico, Camillero, ...) sin que el hospital
// sepa cuál es. Asume que el candado ya está tomado.
func (h *Hospital) despachar(p *Paciente, ubicacion string) (RegistroEpisodio, error) {
	var errores []error
	for i := 0; i < len(h.personal); i++ {
		indice := (h.turno + i) % len(h.personal)
		a := h.personal[indice]

		registro, err := a.Atender(p, ubicacion)
		if err != nil {
			errores = append(errores, fmt.Errorf("%s: %w", a.Nombre(), err))
			continue
		}
		h.turno = (indice + 1) % len(h.personal)
		return registro, nil
	}
	return RegistroEpisodio{}, fmt.Errorf("%w: %w", ErrSinPersonal, errors.Join(errores...))
}

// AsignarHabitacion busca la primera habitación disponible y acuesta allí al
// paciente (estado DormidoEnCama). Si no hay ninguna, retorna un error que
// envuelve a ErrSinHabitacion y el paciente sigue DormidoEnPasillo: no se
// inventa ninguna habitación.
func (h *Hospital) AsignarHabitacion(p *Paciente) (*Habitacion, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.verificarAdmitido(p); err != nil {
		return nil, err
	}
	return h.asignarHabitacion(p)
}

// asignarHabitacion es la versión sin candado, para reutilizarla dentro de
// RegistrarEpisodio (que ya tiene el candado; sync.Mutex no es reentrante y
// tomarlo dos veces bloquearía el programa).
func (h *Hospital) asignarHabitacion(p *Paciente) (*Habitacion, error) {
	if p.Estado() != DormidoEnPasillo {
		return nil, fmt.Errorf("%s no necesita cama (estado: %s)", p.ID(), p.Estado())
	}
	for _, hab := range h.habitaciones {
		if !hab.EstaDisponible() {
			continue
		}
		if err := hab.Ocupar(p); err != nil {
			return nil, err
		}
		p.acostarEn(hab)
		return hab, nil
	}
	return nil, fmt.Errorf("%w: %s permanece en el pasillo (%s)",
		ErrSinHabitacion, p, p.Ubicacion())
}

// DespertarPaciente despierta al paciente y libera su habitación, tomando el
// candado del hospital (versión segura para goroutines de Paciente.Despertar).
func (h *Hospital) DespertarPaciente(p *Paciente) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.verificarAdmitido(p); err != nil {
		return err
	}
	return p.Despertar()
}

// verificarAdmitido asume que el candado ya está tomado.
func (h *Hospital) verificarAdmitido(p *Paciente) error {
	if p == nil {
		return errors.New("paciente nil")
	}
	if h.porID[p.ID()] != p {
		return fmt.Errorf("%w: %s", ErrPacienteNoAdmitido, p.ID())
	}
	return nil
}

// ---------------------------------------------------------------------------
// Consultas (sección 5 del enunciado). Retornan datos, NO imprimen.
// ---------------------------------------------------------------------------

// PacientesEnPasillo (consulta 5.1) retorna los pacientes dormidos en un
// pasillo, en orden de admisión, para despachar a un camillero.
func (h *Hospital) PacientesEnPasillo() []*Paciente {
	h.mu.Lock()
	defer h.mu.Unlock()

	var enPasillo []*Paciente
	for _, p := range h.pacientes {
		if p.Estado() == DormidoEnPasillo {
			enPasillo = append(enPasillo, p)
		}
	}
	return enPasillo
}

// ReporteSeveridad (consulta 5.4) cruza dos fuentes: los pacientes de nivel
// Severo y el historial de episodios de HOY. Retorna cuántos episodios tuvo
// cada paciente severo (0 si no tuvo ninguno).
//
// Se usa un mapa porque la pregunta es "¿cuántos episodios tiene ESTE
// paciente?": la clave es el paciente y el valor su conteo.
func (h *Hospital) ReporteSeveridad() map[*Paciente]int {
	h.mu.Lock()
	defer h.mu.Unlock()

	reporte := make(map[*Paciente]int)
	for _, p := range h.pacientes {
		if p.Nivel() == Severo {
			reporte[p] = 0
		}
	}

	hoy := time.Now()
	for _, r := range h.historial {
		if !mismoDia(r.FechaHora(), hoy) {
			continue
		}
		if _, esSevero := reporte[r.Paciente()]; esSevero {
			reporte[r.Paciente()]++
		}
	}
	return reporte
}

// mismoDia compara solo año, mes y día.
func mismoDia(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// ---------------------------------------------------------------------------
// Consultas auxiliares (retornan copias para proteger el estado interno)
// ---------------------------------------------------------------------------

// Nombre retorna el nombre del hospital.
func (h *Hospital) Nombre() string { return h.nombre }

// Medicos retorna una copia de la lista de médicos.
func (h *Hospital) Medicos() []*Medico {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*Medico(nil), h.medicos...)
}

// Personal retorna una copia del personal que atiende emergencias.
func (h *Hospital) Personal() []Atendedor {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]Atendedor(nil), h.personal...)
}

// Pacientes retorna una copia de la lista de pacientes.
func (h *Hospital) Pacientes() []*Paciente {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*Paciente(nil), h.pacientes...)
}

// Habitaciones retorna una copia de la lista de habitaciones.
func (h *Hospital) Habitaciones() []*Habitacion {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*Habitacion(nil), h.habitaciones...)
}

// Historial retorna una copia de todos los episodios registrados.
func (h *Hospital) Historial() []RegistroEpisodio {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]RegistroEpisodio(nil), h.historial...)
}
