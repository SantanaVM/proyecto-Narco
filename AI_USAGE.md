# ## Herramientas usadas y para qué

- **Claude (Anthropic)**: generó la primera versión completa del proyecto (paquete `hospital`, `main.go`, pruebas, README), la traducción al español del enunciado y una guía para preparar la sustentación. Se le dio como contexto el enunciado en PDF y el tutorial del profesor (https://sebastianzapatar.github.io/tutorialgo/).

Después de recibir el código, lo revisamos archivo por archivo, lo ejecutamos (`go run .`, `go test -race ./...`) y _(go run -race .)

## Prompts decisivos

1. "Necesitamos crear un proyecto en go desde cero, la idea es que me ayudes a entender como puedo integrarlo con un goroutine, y por cierto necesito que me expliques que es eso"
2. "Ya me funciona el proyecto, pero no comprendo porque no me deja correrlo con -race, si me deja correrlo normal"
3. "Necesito hacer una interfaz grafica para el proyecto SIN CAMBIAR NADA DEL CODIGO, hazmo mediante apis ya existentes"

## Casos en que el código generado estaba mal o no era idiomático

1. **Interfaz duplicada de la biblioteca estándar.** En la prueba de los `String()` la IA declaró su propia interfaz `fmtStringer interface{ String() string }`. Eso ya existe en Go como `fmt.Stringer`. Se reemplazó por `fmt.Stringer` y se borró la interfaz repetida.
2. **Salida desalineada.** `Resumen()` usaba un ancho fijo de 20 caracteres para la ubicación, pero "pasillo 2, segundo piso" es más largo y descuadraba las columnas. Se amplió el ancho a 24.
3. **Consulta 5.1 vacía.** Siguiendo el orden de la sección 6 al pie de la letra, cuando se imprimía la consulta del pasillo el paciente P-004 ya tenía cama, y la consulta salía vacía. Se decidió imprimirla también justo después de los ataques.
4. **Interfaz que cambió el código.** Al intentar integrar la interfaz mediante APIs, se tenian que cambiar varias lineas fundamentales de codigo, luego de muchas pruebas se decidió no integrarlo.

## Qué aprendió cada integrante

- **Valentina:** Comprendí de manera mas clara y precisa el uso de go y su aplicación en proyectos reales, al igual que su estructura y funciones adicionales. Me costó demasiado entenderlo al inicio ya que intentaba compararlo todo el tiempo con lenguaje de java, despues de superar esa etapa comencé a sentir que en realidad era mas sencillo de lo que esperaba. Puedo decir, que entendí como programar en Go.
- **Integrante 2:** Aprendi que en realidad GO y java son muy similares en cuanto POO se refiere, entendi bien como funciona su estructura ademas de que entendi mejor como funciona los punteros ademas de tener cosas nuevas como los maps, es un lemguaje no tan complejo si ya tiene experiencia con programacion en java
