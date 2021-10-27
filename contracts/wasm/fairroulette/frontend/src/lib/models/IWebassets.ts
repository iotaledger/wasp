export interface IFoundationData {
  sites: {
    label: string;
    url: string;
  }[];

  footerSections: {
    label: string;
    items: {
      label: string;
      url: string;
    }[];
  }[];

  registeredAddress: {
    label: string;
    value: string[];
  };

  visitingAddress: {
    label: string;
    value: string[];
  };

  information: {
    label: string;
    value?: string;
    urls?: {
      label: string;
      url: string;
    }[];
  }[];
}
