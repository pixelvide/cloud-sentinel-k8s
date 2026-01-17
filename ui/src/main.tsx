import React from "react";
import ReactDOM from "react-dom/client";
import { RouterProvider } from "react-router-dom";
import { router } from "./routes";
import { AppearanceProvider } from "./components/appearance-provider";
import "./index.css";
import "./i18n";

ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
        <AppearanceProvider
            defaultTheme="system"
            defaultColorTheme="default"
            defaultFont="maple"
        >
            <RouterProvider router={router} />
        </AppearanceProvider>
    </React.StrictMode>
);
