package messages

const (
	Welcome                      = "Witaj! Podaj proszę nazwę swojej ulicy."
	AwaitingNumber               = "Zrozumiałem. Jaki jest numer budynku/domu?"
	AwaitingPostcode             = "Dzięki! Jaki jest Twój kod pocztowy?"
	ErrorFindLocation            = "Wystąpił błąd podczas wyszukiwania identyfikatora lokalizacji."
	RegistrationCanceled         = "Rejestracja została anulowana. Wpisz /start, aby spróbować ponownie."
	SomethingWentWrong           = "Przepraszamy, coś poszło nie tak. Wpisz /start, aby rozpocząć proces rejestracji od nowa."
	PleaseStart                  = "Wpisz /start, aby rozpocząć proces rejestracji."
	MultipleLocationsPrompt      = "Znaleziono wiele lokalizacji. Wybierz odpowiednią z poniższej listy:"
	CancelButtonLabel            = "❌ Żadna z powyższych"
	SelectedLocationEdit         = "Wybrana lokalizacja: %s"
	SchedulePrompt               = "Kiedy chcesz otrzymywać powiadomienia? Wybierz wszystkie pasujące opcje, a następnie kliknij Zapisz i zarejestruj poniżej:"
	DoneButtonLabel              = "💾 Zapisz i zarejestruj"
	SaveError                    = "Coś poszło nie tak podczas zapisywania Twojej lokalizacji i preferencji. Wpisz /start, aby spróbować ponownie."
	ConfirmationWithPreferences  = "Wybór lokalizacji potwierdzony (%s)! Rejestracja przebiegła pomyślnie. Będziesz otrzymywać powiadomienia o godzinach: %s."
	ConfirmationNoPreferences    = "Wybór lokalizacji potwierdzony (%s)! Rejestracja przebiegła pomyślnie (nie wybrano godzin powiadomień)."
	AwaitingConfirmationReminder = "Wybierz adres za pomocą przycisków powyżej."
	AwaitingScheduleReminder     = "Wybierz preferencje powiadomień za pomocą przycisków powyżej."

	// Schedule Options
	OptDayBefore19 = "Dzień wcześniej o 19:00"
	OptDayBefore20 = "Dzień wcześniej o 20:00"
	OptDayBefore21 = "Dzień wcześniej o 21:00"
	OptMorning7    = "Rano o 07:00"
	OptMorning8    = "Rano o 08:00"
	OptMorning9    = "Rano o 09:00"
	OptMorning10   = "Rano o 10:00"
)
