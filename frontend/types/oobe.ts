export enum SetupType {
    HOSTNAME = "hostname",
    INPUT = "input",
    SECRET = "secret",
    NUMBER = "number",
    TEXTBOX = "textbox",
    BOOLEAN = "boolean",
    DROPDOWN = "dropdown",
}

export type SetupOption = {
    id: string;
    type: SetupType;
    name: string;
    description: string;
    sticky: boolean;
    must_secure?: boolean;
    regexp?: string;
    required: boolean;
    list_items: {name: string; value: string}[];
};

export class POError extends Error {}
