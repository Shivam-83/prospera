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

    const [salaryData, setSalaryData] = useState({
        currentSalary: 0,
        medianSalary: 0,
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
                return;
            }

            try {
                console.log("Fetching salary data for userId:", storedUserId);
                const response = await axios.get(
                    `${API_BASE}/salary/benchmark?userId=${storedUserId}`
                );
                const data = response.data;

                // Update state with data from the server
                setSalaryData({
                    currentSalary: data.CurrentSalary,
                    medianSalary: data.DesiredSalary,
                    minSalary: data.minSalary || 45000,
                    maxSalary: data.maxSalary || 70000,
                    jobTitle: data.jobTitle,
                    location: data.Location,
                    yearsExperience: data.YearsExperience,
                });
            } catch (error) {
                console.error("Error fetching salary data:", error);
            }
        };

        fetchSalaryData();
    }, [API_BASE]);

    // Data for the salary chart
    const graphData = [
        { name: "Current", value: salaryData.currentSalary },
        { name: "Median", value: salaryData.medianSalary },
        { name: "Min", value: salaryData.minSalary },
        { name: "Max", value: salaryData.maxSalary },
    ];

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
                    { title: "Median Industry Salary", value: `${salaryData.medianSalary}€`, imgSrc: median },
                    { title: "Min Industry Salary", value: `${salaryData.minSalary}€`, imgSrc: min },
                    { title: "Max Industry Salary", value: `${salaryData.maxSalary}€`, imgSrc: max },
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
