import { Outlet } from "react-router-dom";
import { ClientLayout } from "./components/ClientLayout";
import { Toaster } from "sonner";

function App() {
    return (
        <ClientLayout>
            <Outlet />
            <Toaster position="bottom-left" richColors />
        </ClientLayout>
    );
}

export default App;
