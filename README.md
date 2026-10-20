# Hospital de los Costeños con Narcolepsia

Sistema de gestión del hospital escrito en Go (solo biblioteca estándar).
Programación Orientada a Objetos, Universidad EIA. Profesor: Sebastián Zapata Ramírez.

## Integrantes

- Valentina Santana
- Isabella Ramírez

## Cómo ejecutar

Requiere Go 1.21 o superior.

```bash
go run .                   # escenario obligatorio de la sección 6
go run . -simular          # escenario + simulación concurrente (bono)
go run -race . -simular    # lo mismo, con el detector de carreras
```

## Cómo correr las pruebas

```bash
go test ./...              # todas las pruebas
go test -race ./...        # con el detector de carreras
go vet ./...               # no debe reportar nada
gofmt -l .                 # no debe listar ningún archivo
```

## Estructura

```
Proyecto GO/
  go.mod                  módulo "proyectogo"
  main.go                 solo arma el escenario e imprime
  hospital/
    hospital.go           Hospital: registro, episodios y las 4 consultas
    persona.go            Persona (id, nombre, edad), embebida en los demás
    paciente.go           Paciente
    medico.go             Medico
    camillero.go          Camillero (segundo tipo que cumple Atendedor)
    atendedor.go          interfaz Atendedor
    habitacion.go         Habitacion
    episodio.go           RegistroEpisodio
    estados.go            constantes tipadas (estado, nivel, estado de habitación)
    simulacion.go         bono: una goroutine por paciente + sync.Mutex
    hospital_test.go      10 pruebas
```

## Correspondencia con el enunciado

El enunciado permite nombres en inglés o en español si son consistentes. Usamos español en todo el proyecto:

| Enunciado (inglés)            | Proyecto (español)                         |
|-------------------------------|--------------------------------------------|
| `Hospital`                    | `Hospital`                                 |
| `Patient`                     | `Paciente`                                 |
| `Doctor`                      | `Medico`                                   |
| `Room`                        | `Habitacion`                               |
| `EpisodeRecord`               | `RegistroEpisodio`                         |
| `Person`                      | `Persona`                                  |
| `Attender`                    | `Atendedor`                                |
| `Orderly`                     | `Camillero`                                |
| `AdmitPatient` / `HireDoctor` | `AdmitirPaciente` / `ContratarMedico`      |
| `AssignRoom`                  | `AsignarHabitacion`                        |
| `RegisterEpisode`             | `RegistrarEpisodio`                        |
| `SufferSleepAttack` / `WakeUp`| `SufrirAtaqueDeSueno` / `Despertar`        |
| `DiagnosePatient`             | `DiagnosticarPaciente`                     |
| `AttendSleepEmergency`        | `AtenderEmergenciaDeSueno`                 |
| `IsAvailable` / `Occupy` / `Release` | `EstaDisponible` / `Ocupar` / `Liberar` |
| `Summary`                     | `Resumen`                                  |
| 5.1 `PatientsInHallway`       | `PacientesEnPasillo() []*Paciente`         |
| 5.2 `MyEpisodes`              | `MisEpisodios() []RegistroEpisodio`        |
| 5.3 `AssignRoom`              | `AsignarHabitacion(p) (*Habitacion, error)`|
| 5.4 `SevereReport`            | `ReporteSeveridad() map[*Paciente]int`     |
| `Awake / AsleepInHallway / AsleepInBed` | `Despierto / DormidoEnPasillo / DormidoEnCama` |
| `Mild / Moderate / Severe`    | `Leve / Moderado / Severo`                 |
| `Available / Occupied`        | `Disponible / Ocupada`                     |

## Decisiones de diseño

- **Encapsulamiento.** Todos los campos de las entidades están en minúscula. Desde `main` (otro paquete) solo se puede leer o cambiar el estado a través de métodos. Los métodos que retornan listas (`Ocupantes`, `MisEpisodios`, `Pacientes`, ...) devuelven una **copia**, para que nadie modifique la lista interna desde fuera.
- **Composición.** `Persona` está embebida **por valor** en `Paciente`, `Medico` y `Camillero`. Sus métodos `ID()`, `Nombre()` y `Edad()` se promueven, y además son justo dos de los tres métodos de `Atendedor`.
- **Polimorfismo.** El hospital guarda su personal en un `[]Atendedor` (médicos y camillero juntos). `despachar` recorre esa lista por turnos (round-robin) y llama `Atender(...)` sin saber qué tipo concreto es. Si alguien falla, pasa al siguiente. `ContratarPersonal(a Atendedor)` permite agregar tipos nuevos (por ejemplo una enfermera) **sin cambiar el paquete**.
- **Médico vs. camillero.** El médico diagnostica (se vuelve el médico a cargo del paciente) y guarda sus episodios; el camillero solo recoge al paciente y no diagnostica.
- **Relación médico → episodios.** Cada médico guarda sus propios episodios, por eso `MisEpisodios()` no recibe el hospital.
- **Errores.** Todo lo que puede fallar retorna `error` como último valor; no se usa `panic`. Los errores esperables son centinelas (`ErrSinHabitacion`, `ErrPacienteNoAdmitido`, `ErrSinPersonal`) envueltos con `%w`, para poder reconocerlos con `errors.Is`.
- **Sin cama.** Si no hay habitación, `RegistrarEpisodio` igual registra el episodio (el paciente fue atendido en el pasillo), el paciente queda `DormidoEnPasillo` y se retorna un error que envuelve `ErrSinHabitacion`. Nunca se crea una habitación nueva.
- **Slice + mapa para pacientes.** El slice conserva el orden de admisión (útil para listar); el mapa `porID` permite detectar ids repetidos en tiempo constante.
- **Mapa en el reporte de severidad.** La pregunta es "¿cuántos episodios tuvo *este* paciente?", así que la clave es el paciente y el valor el conteo. Como los mapas no tienen orden, `main` ordena por id antes de imprimir.
- **Concurrencia.** Un solo `sync.Mutex` en el `Hospital` protege todo el estado. Todos los métodos exportados del hospital lo toman. Como `sync.Mutex` no es reentrante, hay versiones internas sin candado (`asignarHabitacion`) que se reutilizan desde métodos que ya lo tienen. Los métodos de las entidades (`Paciente.Despertar`, etc.) no tienen candado propio: en código concurrente se deben usar a través del hospital (`DespertarPaciente`).
- **Consulta 5.1 dos veces.** En el escenario la consulta del pasillo se imprime tras los ataques (aparece P-004) y otra vez tras reasignar la habitación (queda vacía), para que el resultado se vea en los dos momentos.
