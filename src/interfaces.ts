export interface SimpleUser {
    id: string;
    name: string;
    teacher: boolean;
    children: string[];
    type: string;
};

export interface Receiver {
    id: string;
    name: string;
};

export interface School {
    id: string;
    name: string;
    logo: string;
};

export interface Grade {
    id: string;
    fullName: string;
    subject: string;
    deliberation: string;
    conceptual: string;
    averageFirstFour: string;
    averageSecondFour: string;
    final: string;
}

export interface GradeStudent {
    subject: string;
    deliberation: string;
    conceptual: string;
    averageFirstFour: string;
    averageSecondFour: string;
    final: string;
}

export interface GradeParsed {
    deliberation: string;
    conceptual: string;
    averageFirstFour: string;
    averageSecondFour: string;
    final: string;
}

export interface User {
    token?: string;
    id: string;
    name: string;
    schoolName: string;
    schoolLogo?: string;
    email: string;
    password?: string;
    verified?: boolean;
    verificator?: string;
    tfa: string | boolean;
    administrator?: string[];
    teacher?: string;
    parents?: string[];
    pendingparents?: string[];
    children?: string[];
    grades?: Grade[] | GradeStudent[] | Grade[][];
    avaliable: SimpleUser[];
    schools?: School[];
    pendingschools?: School[];
};

export interface File {
    id: string;
    name: string;
};