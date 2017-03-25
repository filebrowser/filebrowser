class Actividad {
	method calcularMejora()
}

class EstudiarMateria inherits Actividad {
	var materia
	var puntos = 0
	
	new(m, p) {
		materia = m
		puntos = p
	}
	
	override method calcularMejora() = puntos
}

class EjercitarEnSimulador inherits Actividad {
	var horas = 0
	new(h) { horas = h }
	override method calcularMejora() = 10 * horas
}

object pepita {
	var energia = 100
	method volar(m) {
		energia -= m
	}
}