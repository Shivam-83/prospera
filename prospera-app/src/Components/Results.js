import React, { useState, useEffect } from "react";
import { Link, useSearchParams } from "react-router-dom";
import axios from "axios";
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    Legend,
    ResponsiveContainer,
} from "recharts";
import salary from "../Assets/salary.png";
import max from "../Assets/max.png";
import min from "../Assets/min.png";
import median from "../Assets/Median.png";

const Results = () => {
    const [searchParams] = useSearchParams();
    const API_BASE = process.env.REACT_APP_API_URL || "https://prospera-bnny.onrender.com";

    const [isLoading, setIsLoading] = useState(true);
    const [salaryData, setSalaryData] = useState({
        currentSalary: 0,
        desiredSalary: 0,
        minSalary: 0,
        maxSalary: 0,
        jobTitle: "",
        location: "",
        yearsExperience: 0,
    });

    useEffect(() => {
        const fetchSalaryData = async () => {
            const storedUserId = localStorage.getItem("userId");

            if (!storedUserId) {
                console.warn("userId not found in localStorage");
                alert("Session expired. Please fill the form again.");
                window.location.href = "/input-form";
                return;
            }

            try {
                console.log("Fetching salary data for userId:", storedUserId);
                const response = await axios.get(
                    `${API_BASE}/salary/benchmark?userId=${storedUserId}`
                );
                const data = response.data;

                setSalaryData({
                    currentSalary: data.CurrentSalary,
                    desiredSalary: data.DesiredSalary,
                    // minSalary and maxSalary are not returned by the backend yet;
                    // they will be 0 until the backend computes real market ranges.
                    minSalary: data.minSalary || 0,
                    maxSalary: data.maxSalary || 0,
                    jobTitle: data.jobTitle,
                    location: data.Location,
                    yearsExperience: data.YearsExperience,
                });
            } catch (error) {
                console.error("Error fetching salary data:", error);
                if (error.response?.status === 404) {
                    alert(
                        "Session data not found on the server. " +
                        "This usually happens after a server restart. " +
                        "Please fill the form again to create a new session."
                    );
                    localStorage.removeItem("userId");
                    window.location.href = "/input-form";
                } else {
                    alert("Error loading salary data. Please try again.");
                }
            } finally {
                setIsLoading(false);
            }
        };

        fetchSalaryData();
    }, [API_BASE]);

    // Data for the salary chart
    const graphData = [
        { name: "Current", value: salaryData.currentSalary },
        { name: "Target", value: salaryData.desiredSalary },
        { name: "Min", value: salaryData.minSalary },
        { name: "Max", value: salaryData.maxSalary },
    ];

    if (isLoading) {
        return (
            <div style={{ backgroundColor: "#ffeecd", padding: "20px", textAlign: "center", minHeight: "60vh", display: "flex", alignItems: "center", justifyContent: "center" }}>
                <p style={{ fontSize: "1.2rem", color: "#696666" }}>⏳ Loading your salary data...</p>
            </div>
        );
    }

    return (
        <div style={{ backgroundColor: "#ffeecd", padding: "20px" }}>
            {/* Title */}
            <h2
                style={{
                    color: "#696666",
                    fontFamily: "serif",
                    fontSize: "2rem",
                    textAlign: "center",
                }}
            >
                Salary Benchmark
            </h2>

            {/* Horizontal line */}
            <hr style={{ border: "1px solid #696666", margin: "20px 0" }} />

            {/* Salary Data Sections */}
            <div className="salary-boxes-container">
                {[
                    { title: "Current Salary", value: `${salaryData.currentSalary}€`, imgSrc: salary },
                    { title: "Your Target Salary", value: `${salaryData.desiredSalary}€`, imgSrc: median },
                    { title: "Min Industry Salary", value: salaryData.minSalary ? `${salaryData.minSalary}€` : "Ask your AI coach", imgSrc: min },
                    { title: "Max Industry Salary", value: salaryData.maxSalary ? `${salaryData.maxSalary}€` : "Ask your AI coach", imgSrc: max },
                ].map((item, index) => (
                    <div className="salary-box" key={index}>
                        <div className="salary-box-image">
                            <img src={item.imgSrc} alt={item.title} />
                        </div>
                        <div className="salary-box-content">
                            <div className="salary-title">{item.title}</div>
                            <div className="salary-value">{item.value}</div>
                        </div>
                    </div>
                ))}
            </div>

            <br />
            <br />

            {/* Chart Title */}
            <h3 className="chart-title">
                Salaries for {salaryData.jobTitle} in {salaryData.location} with{" "}
                {salaryData.yearsExperience} Years of Experience
            </h3>
            <br />

            {/* Chart */}
            <ResponsiveContainer width="100%" height={300}>
                <LineChart data={graphData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Line type="monotone" dataKey="value" stroke="#8884d8" />
                </LineChart>
            </ResponsiveContainer>

            {/* Confidence Boosting Text */}
            <div style={{ textAlign: "center", margin: "20px 0" }}>
                <p style={{ fontSize: "1.2rem", color: "#696666" }}>
                    Want to practice to boost your confidence & earn what you deserve?
                </p>
            </div>

            {/* Try Again Button */}
            <div className="try-again-button">
                <Link to="/input-form">
                    <button>Try Again</button>
                </Link>
            </div>
        </div>
    );
};

export default Results;
