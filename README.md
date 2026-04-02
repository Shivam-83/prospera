# **Prospera** - Empowering Women in Career Negotiation with AI

Prospera is an AI-driven application designed to help women confidently negotiate salaries and achieve financial equity. Leveraging Google’s Gemini AI, Prospera provides real-time salary benchmarking, personalized negotiation coaching, and career guidance tailored to individual needs and aspirations.

![home](assets/1_home.jpg)

![salary benchmark](assets/3_salary_benchmark_chat.png)

![salary results](assets/6_results.png)
## **Table of Contents**

1. [Prerequisites](#prerequisites)
2. [Installation and Setup](#installation-and-setup)
3. [Usage](#usage)
4. [Testing Prospera](#testing-prospera)
5. [Project Structure](#project-structure)
6. [Built With](#built-with)
7. [Contributing](#contributing)
8. [License](#license)

---
## **Prerequisites**

To get started with Prospera, make sure you have the following installed based on how you want to run the application:

**For all setups:**
1. **Google API Key**: Have an [OpenRouter API Key](https://openrouter.ai/) ready to add it to the `.env` file (see below).
2. **Make**: (Optional but recommended) Useful for running automated commands.

**Option A: Running with Docker (Easiest)**
1. **Docker**: [Get Docker](https://docs.docker.com/get-docker/)
2. **Docker Compose**: Verify with `docker-compose --version`

**Option B: Running Locally (For live development/changes)**
1. **Go (Golang)**: Version 1.23 or newer. [Download Go](https://go.dev/dl/)
2. **Node.js**: Version 18 or newer. [Download Node.js](https://nodejs.org/en)



## **Installation and Setup**

### 1. Clone the Repository

Clone the repository and navigate into the project directory:

```bash
git clone https://github.com/Shivam-83/prospera.git
cd prospera
```

### 2. Configure Environment Variables

Prospera requires specific environment variables for API keys and configuration. Follow these steps to set up your environment:

1. Create a `.env` file in the project root:

    ```bash
    cp .env.example .env
    ```

2. Open `.env` and add your OpenRouter (Google API) key:
    ```
    GOOGLE_API_KEY=your-api-key-here
    ```

    You can obtain an API key from [OpenRouter](https://openrouter.ai/).

3. Create the frontend `.env` file in the `prospera-app` directory:
    
    ```bash
    cd prospera-app
    echo "REACT_APP_API_URL=http://localhost:8080" > .env
    cd ..
    ```

### 3. Running Prospera

### Option A: Run with Docker (Recommended)

To build and run both backend and frontend containers:

```bash
make docker-up
```

To stop the containers:

```bash
make docker-down
```

### Option B: Run Locally Without Docker (Go & Node.js required)

If you are developing locally without Docker, you need to install the dependencies for the frontend first.

1. **Install Frontend Dependencies**:
    
    ```bash
    cd prospera-app
    npm install
    cd ..
    ```

2. **Start the Backend**:

    ```bash
    make run-backend
    ```
    *(If you don't have Make installed, you can simply run `cd app && go run main.go`)*

3. **Start the Frontend**:

    ```bash
    make run-frontend
    ```
    *(If you don't have Make installed, you can simply run `cd prospera-app && npm start`)*

4. **Stopping Services Locally**:

    ```bash
    make kill-local
    ```

---

## **Usage**

Once Prospera is running, you can access the application through your browser:

- **Frontend**: Visit `http://localhost:3000`
- **Backend API**: Available at `http://localhost:8080`

### Features to Explore

- **Salary Benchmarking**: Fill out the input form to receive real-time salary insights.

![input](assets/2_input_information.png)
![input](assets/3_salary_benchmark_chat.png)
![input](assets/6_results.png)

- **Negotiation Coach Chat**: Engage in a real-time chat to practice negotiation scenarios. 
![input](assets/4_negotiation_coach_chat.png)

- **Tips Chat**: Interact with Prospera’s AI-driven coach for personalized advice.
![input](assets/5_boost_confidence_chat.png)

---

## **Testing Prospera**

### Step-by-Step Testing Guide

1. **Input Form**: Use the form on the frontend to input salary details, experience, location, and more. Submit the form to get a salary benchmarking response.
2. **Negotiation Coaching**: Access the negotiation coach to receive interactive guidance. The backend processes responses in real-time using WebSocket, simulating a live conversation.
3. **Chat Interface**: Test the WebSocket functionality by engaging in a back-and-forth conversation with the AI, ensuring seamless real-time responses.

### Viewing Logs

- **Docker logs**: Use `docker-compose logs -f` to monitor backend and frontend logs while testing.
- **Local logs**: When running locally, logs will appear in each terminal window where the backend and frontend are running.

---

## **Project Structure**

Here’s an overview of the main files and directories:

```
plaintext
Copier le code
prospera/
├── app/                   # Go backend
│   ├── main.go            # Backend entry point
│   ├── Dockerfile         # Backend Dockerfile
│   └── ...
├── prospera-app/          # React frontend
│   ├── src/               # React source code
│   ├── Dockerfile         # Frontend Dockerfile
│   └── ...
├── docker-compose.yml     # Docker Compose configuration
├── Makefile               # Makefile for easy setup and testing
├── .env.example           # Example environment file
└── README.md              # Project documentation

```
The full architecture looks like this
![input](assets/architecture_overview.png)

---

## **Built With**

![input](assets/built_with_tech.png)

- **Go**: Backend API and WebSocket server
- **React**: Frontend UI
- **Google Gemini AI**: AI-driven negotiation assistance
- **Docker & Docker Compose**: Containerized environment for easy deployment
- **Make**: Simplifies running and managing local and Docker environments

---

## **Contributing**

Contributions are welcome! Please follow these steps if you would like to contribute to Prospera:

1. Fork the repository
2. Create a new feature branch (`git checkout -b feature/YourFeature`)
3. Commit your changes (`git commit -m 'Add new feature'`)
4. Push to the branch (`git push origin feature/YourFeature`)
5. Open a Pull Request

---

## **License**

This project is licensed under the MIT License - see the LICENSE file for details.

---

## **Contact**

For any questions or feedback, please open an issue.
