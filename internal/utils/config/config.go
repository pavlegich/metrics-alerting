// Пакет config содержит методы для
// работы с конфигурационными данными
package config

import "flag"

func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
