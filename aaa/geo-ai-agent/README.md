<p align="center">
  <a href="https://brightdata.com/">
    <img src="https://mintlify.s3.us-west-1.amazonaws.com/brightdata/logo/light.svg" width="300" alt="Bright Data Logo">
  </a>
</p>

<div align="center">
  <img src="https://img.shields.io/badge/python-3.10+-blue"/>
  <img src="https://img.shields.io/badge/License-MIT-blue"/>
</div>

---

# 🚀 GEO AI Crew

GEO Agent Crew uses [CrewAI](https://crewai.com) to automate AI-driven webpage content audits. Enter a URL, and the system accesses the webpage, extracts its title, generates and summarizes related queries using [Gemini with the Google Search tool](https://ai.google.dev/gemini-api/docs/google-search), fetches Google AI Overviews via [Bright Data SERP API](https://brightdata.com/products/serp-api), compares results, and outputs actionable page-level optimization suggestions in Markdown file.

<img src="https://github.com/brightdata/geo-ai-agent/blob/main/GEO%20diagram.png"/>

---

## 🤖 Understanding Your Crew

The `ai-content-optimization-agent` Crew is composed of six AI agents, each with unique roles, goals, and tools. These agents collaborate on a series of tasks, defined in `config/tasks.yaml`, leveraging their collective skills to achieve complex objectives. The `config/agents.yaml` file outlines the capabilities and configurations of each agent in your crew.

## 🛠️ Installation

Ensure you have **Python >=3.10 <3.14** installed on your system.

This project uses [`uv`](https://docs.astral.sh/uv/) for dependency management and package handling.
First, if you haven't already, install `uv`:

```bash
pip install uv
```

Next, navigate to your project directory and install the project's dependencies:

```bash
cd geo-ai-agent
uv sync
```

---

## 🔑 Environment Configuration

This project requires four environment variables to work:
- **`GEMINI_API_KEY`**: Your Gemini API key.
- **`MODEL`**: The name of the Gemini model to power your crew of agents (e.g., `gemini/gemini-2.5-flash`).
- **`BRIGHT_DATA_API_KEY`**: Your [Bright Data API key](https://docs.brightdata.com/api-reference/authentication).
- **`BRIGHT_DATA_ZONE`**: The name of the [Web Unlocker zone in your Bright Data dashboard](https://docs.brightdata.com/scraping-automation/web-unlocker/quickstart) you want to connect to.

Define them directly in your terminal or place them in a `.env` file at the root of your project:
```
geo-ai-agent/
├── ...
├── .env # <---
└── src/
    └── ai_content_optimization_agent/
        └── ...
```
Populate the `.env` file like this:
```
GEMINI_API_KEY="<YOUR_GEMINI_API_KEY>"
MODEL="<CHOSEN_GEMINI_MODEL>"
BRIGHT_DATA_API_KEY="<BRIGHT_DATA_API_KEY>"
BRIGHT_DATA_ZONE="<YOUR_BRIGHT_DATA_ZONE>"
```

## ▶️ Running the Project
Activate the `.venv` created by the `uv sync` command:
```bash
 source .venv/bin/activate
```
Or, on Windows:
```powershell
.venv/Scripts/activate
```

With the virtual environment activated, start your crew of AI agents by running the following command from the root folder of your project:

```bash
crewai run
```

This command initializes the `ai-content-optimization-agent` crew, assembling the agents and assigning them tasks as defined in the CrewAI configuration files.

☑️ This application will produce a `output/report.md` file, along with other `ouput/*.md` files containing intermediate data and results from the agents.

---

### ⚙️ Customizing
- 🔧 Update the `MODEL` environment variable to change the Gemini model used by this crew of agents.
- 🧑‍💻 Edit `src/ai_content_optimization_agent/config/agents.yaml` to modify the definitions of the agents. 
- 📋 Edit `src/ai_content_optimization_agent/config/tasks.yaml` to modify the definitions of the tasks assigned to the agents. 
- 🛠️ Update `src/ai_content_optimization_agent/crew.py` to integrate other AI models or add your own logic and tools.
- ⚡ Edit `src/ai_content_optimization_agent/main.py` to add custom inputs for your agents and tasks.

---

## 💬 Support

For support, questions, or feedback regarding the `ai-content-optimization-agent` Crew or CrewAI:

- ☀️ Visit Bright Data's [SERP API docs](https://docs.brightdata.com/scraping-automation/serp-api/introduction)
- 📖 Visit CrewAI's [documentation](https://docs.crewai.com)
- 🐙 Reach out to CrewAI through the [GitHub repository](https://github.com/joaomdmoura/crewai)
- 💬 [Join Discord](https://discord.com/invite/X4JWnZnxPb)
- 💡 [Chat with CrewAI's docs](https://chatg.pt/DWjSBZn)

---

✨ Let's create wonders together with the power and simplicity of Bright Data & CrewAI.
