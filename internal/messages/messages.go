package messages

const (
	Welcome = "Witaj! \n\n" +
		"ℹ️ Ta aplikacja nie jest oficjalną aplikacją miasta Warszawa.\n" +
		"🔒 Dla Twojego bezpieczeństwa nie zapisujemy Twojego numeru telefonu ani nazwy użytkownika. " +
		"Zapisywany jest wyłącznie identyfikator chatu (ChatID) oraz wybrany adres.\n" +
		"Wpisz /prywatnosc, aby przeczytać Politykę Prywatności.\n\n" +
		"Jeśli chcesz całkowicie usunąć swoje dane, w dowolnym momencie wpisz /anuluj. \n\n" +
		"Aby zacząć, podaj proszę nazwę swojej ulicy:"
	AwaitingNumber               = "Zrozumiałem. Jaki jest numer budynku/domu?"
	AwaitingPostcode             = "Dzięki! Jaki jest Twój kod pocztowy?"
	ErrorFindLocation            = "Wystąpił błąd podczas wyszukiwania identyfikatora lokalizacji."
	RegistrationCanceled         = "Rejestracja została anulowana. Wpisz /start, aby spróbować ponownie."
	SomethingWentWrong           = "Przepraszamy, coś poszło nie tak. Wpisz /start, aby rozpocząć proces rejestracji od nowa."
	UnknownCommand               = "Nie rozumiem tej komendy lub wiadomości. Dostępne komendy:\n/start - Rozpocznij rejestrację\n/harmonogram - Sprawdź najbliższe wywozy\n/anuluj - Usuń swoje dane"
	MultipleLocationsPrompt      = "Znaleziono wiele lokalizacji. Wybierz odpowiednią z poniższej listy:"
	CancelButtonLabel            = "❌ Żadna z powyższych"
	SelectedLocationEdit         = "Wybrana lokalizacja: %s"
	SchedulePrompt               = "Kiedy chcesz otrzymywać powiadomienia? Wybierz z poniższej listy lub wpisz własną godzinę (np. '6' lub '6:00' dla powiadomienia o 6 rano w dniu wywozu, 'W 15' dla powiadomienia dzień wcześniej o 15:00):"
	DoneButtonLabel              = "💾 Zapisz i zarejestruj"
	SaveError                    = "Coś poszło nie tak podczas zapisywania Twojej lokalizacji i preferencji. Wpisz /start, aby spróbować ponownie."
	ConfirmationWithPreferences  = "Wybór lokalizacji potwierdzony (%s)! Rejestracja przebiegła pomyślnie. Będziesz otrzymywać powiadomienia o godzinach: %s.\n\nTwój harmonogram wywozu będzie odświeżany automatycznie z serwerów miasta co 3 dni.\n\nDostępne komendy:\n/harmonogram - Sprawdź wywozy\n/anuluj - Usuń dane\n/start - Rejestracja od nowa"
	ConfirmationNoPreferences    = "Wybór lokalizacji potwierdzony (%s)! Rejestracja przebiegła pomyślnie (nie wybrano godzin powiadomień).\n\nTwój harmonogram wywozu będzie odświeżany automatycznie z serwerów miasta co 3 dni.\n\nDostępne komendy:\n/harmonogram - Sprawdź wywozy\n/anuluj - Usuń dane\n/start - Rejestracja od nowa"
	AwaitingConfirmationReminder = "Wybierz adres za pomocą przycisków powyżej."
	AwaitingScheduleReminder     = "Wybierz preferencje powiadomień za pomocą przycisków powyżej."
	DataDeleted                  = "Twoje dane zostały usunięte z systemu. Jeśli chcesz zacząć od nowa, wpisz /start."

	PrivacyPolicy = `🛡️ Polityka Prywatności

1. Administrator danych: Twórca tego bota.
2. Cel przetwarzania: Wysyłanie automatycznych powiadomień o harmonogramie wywozu śmieci.
3. Zbierane dane:
   - Telegram Chat ID (Twoje imię, nazwa użytkownika czy numer telefonu z Telegrama NIE są zapisywane).
   - Wybrany adres (dokładna nazwa lokalizacji oraz identyfikator punktu wywozu zwrócone przez miejskie API Warszawy).
4. Przechowywanie danych: Dane są przechowywane w bazie wyłącznie po to, aby móc wysyłać powiadomienia. W każdej chwili możesz je usunąć za pomocą komendy /anuluj.
5. Status: Aplikacja jest projektem prywatnym i nie jest oficjalnym produktem m.st. Warszawy. Dane nie są przekazywane podmiotom trzecim.`

	HarmonogramError         = "Wystąpił błąd podczas pobierania harmonogramu. Spróbuj ponownie później."
	HarmonogramNotRegistered = "Nie znaleziono Twojego adresu w bazie. Wpisz /start aby rozpocząć rejestrację."
	HarmonogramPending       = "Jesteś zarejestrowany, ale Twój harmonogram nie został jeszcze pobrany z systemu. Spróbuj ponownie za jakiś czas."
	HarmonogramHeaderDate    = "📅 Harmonogram (aktualizacja z bazy: %s)\n\n"
	HarmonogramHeaderNoDate  = "📅 Harmonogram (brak danych o aktualizacji)\n\n"
	HarmonogramNoData        = "Brak danych "

	FractionZmieszane        = "Zmieszane"
	FractionPapier           = "Papier"
	FractionPlastik          = "Plastik"
	FractionSzklo            = "Szkło"
	FractionBio              = "Bio"
	FractionZielone          = "Zielone"
	FractionBioRestauracyjne = "Bio Restauracyjne"
	FractionGabaryty         = "Gabaryty"

	// Schedule Options
	OptDayBefore19 = "Dzień wcześniej o 19:00"
	OptDayBefore20 = "Dzień wcześniej o 20:00"
	OptDayBefore21 = "Dzień wcześniej o 21:00"
	OptMorning7    = "Rano o 07:00"
	OptMorning8    = "Rano o 08:00"
	OptMorning9    = "Rano o 09:00"
	OptMorning10   = "Rano o 10:00"
)
