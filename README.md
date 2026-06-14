# Śmieci SMS (Warsaw Garbage Collection Reminder Bot)

## 📌 What is this project?
This project is an automated bot system designed to remind residents about upcoming garbage collection days. Users can register their address via a Telegram bot, select their preferred notification times (e.g., the day before at 19:00, or the morning of at 07:00), and the system automatically fetches the latest garbage schedule from the city's database. When the collection day approaches, users receive timely reminders.

**Try it out on Telegram!**
🤖 Name: **Śmieci Warszawa powiadomienia**
🔗 ID: [@smieciwarszawapowiadomienia_bot](https://t.me/smieciwarszawapowiadomienia_bot)

## 🎯 Why was it created?
In Warsaw, the garbage operator comes very frequently—on average every 2 days, and sometimes even two/three days in a row! To make matters more complicated, each day is dedicated to a completely different type of waste (e.g., paper one day, plastics the next). The schedule also likes to change quite often. 

Usually, residents have to manually go to the city's website and look up their street to check what needs to be taken out. Doing this regularly is incredibly annoying and easy to forget. This project solves that problem by automatically tracking the city's updates and sending you a simple ping right when you need it.

Additionally, this project serves as a personal learning ground. I am using this repository to learn and experiment with:
- **Go (Golang)**: Building robust backend services, working with APIs, and handling concurrency.
- **Docker**: Containerizing the application and PostgreSQL database for consistent local development and deployment.
- **Terraform**: Provisioning and managing infrastructure as code (IaC) for deployments.
- **GitHub Actions**: Automating CI/CD pipelines to easily test, build, and deploy the application.
- **Antigravity**: Leveraging advanced AI agentic tools for pair programming, debugging, and development automation.

## 🛠 Key Features
- Interactive Telegram bot for easy user registration (`/start`, `/harmonogram`, `/anuluj`).
- Automated data fetching and synchronization with external municipal garbage schedules (updates every 3 days).
- Background jobs for processing schedules and sending out notifications reliably.
- Persistent and safe data storage using PostgreSQL.
- Fully containerized environment orchestratable via `docker-compose`.

---
*Project in progress.*
