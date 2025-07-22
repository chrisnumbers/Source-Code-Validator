// client/src/App.js or App.tsx
import React, { useState } from "react";
import axios from "axios";

function App() {
  const [url, setUrl] = useState("");
  const [file, setFile] = useState(null);
  const [result, setResult] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();

    const formData = new FormData();
    formData.append("url", url);
    if (file) formData.append("requirements", file);

    try {
      const res = await axios.post("http://localhost:8080/validate", formData);
      setResult(res.data.message || "Success");
    } catch (error) {
      console.error(error);
      setResult("Validation failed or server error.");
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-6">
      <div className="max-w-2xl w-full bg-white shadow-xl rounded-2xl p-8">
        <h1 className="text-3xl font-bold text-blue-600 mb-4">
          Source Code Validator
        </h1>
        <p className="text-gray-600 mb-6">
          Paste your GitHub URL or upload a file to validate your source code.
        </p>

        <form className="space-y-4">
          <input
            type="text"
            placeholder="Enter GitHub Repo URL"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />

          <input type="file" className="w-full text-gray-600" />

          <button
            type="submit"
            className="bg-blue-600 hover:bg-blue-700 text-white font-semibold px-6 py-2 rounded-lg transition"
          >
            Validate
          </button>
        </form>
      </div>
    </div>
  );
}

export default App;
