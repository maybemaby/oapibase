export function FormGroup({ label, htmlFor, children }: { label: string, htmlFor: string, children: React.ReactNode }) {
    return (
        <div className="flex flex-col gap-1">
            <label htmlFor={htmlFor}>{label}</label>
            {children}
        </div >
    )
}