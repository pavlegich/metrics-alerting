// Пакет staticlint содержит набор статических анализаторов исходного кода,
// объединённых в multichecker.
//
// multichecker включает в себя:
//
// - все анализаторы пакета golang.org/x/tools/go/analysis/passes;
//
// - все анализаторы класса SA пакета staticcheck.io;
//
// - анализаторы S1020, ST1003, QF1003 пакета staticcheck.io;
//
// - анализатор корректного закрытия тела запроса request.Body (github.com/timakin/bodyclose/passes/bodyclose);
//
// - анализатор корректного обёртывания ошибок (github.com/fatih/errwrap);
//
// - собственный анализатор предотвращения использования вызова os.Exit в main() пакета main.
//
// Примеры запуска бинарного файла:
//
//	./staticlint -S1020 <путь>
//	./staticlint -fieldalignment <путь>
//	./staticlint -errwrap <путь>
//
// Анализатор предотвращения использования вызова os.Exit в функции main пакета main
// запускается следующей командой:
//
//	./staticlint -exitcheck <путь>
package main
