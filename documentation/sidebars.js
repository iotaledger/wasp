/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

module.exports = {
    // By default, Docusaurus generates a sidebar from the docs folder structure
    //tutorialSidebar: [{type: 'autogenerated', dirName: '.'}],

    // But you can create a sidebar manually
    tutorialSidebar: [
        {
            type: 'doc',
            label: 'Overview',
            id: 'overview',
        },
        {
            type: 'category',
            label: 'Core Concepts',
            items: [
                {
                    type: 'doc',
                    label: 'Smart Contracts',
                    id: 'guide/core_concepts/smart-contracts',
                },
                {
                    type: 'doc',
                    label: 'ISC Architecture',
                    id: 'guide/core_concepts/isc-architecture',
                },
                {
                    type: 'doc',
                    label: 'Validators and Access Nodes',
                    id: 'guide/core_concepts/validators',
                },
                {
                    type: 'doc',
                    label: 'Consensus',
                    id: 'guide/core_concepts/consensus',
                },
                {
                    type: 'doc',
                    label: 'State, Transitions and State Anchoring',
                    id: 'guide/core_concepts/states',
                },
                {
                    type: 'doc',
                    label: 'Anatomy of a Smart Contract',
                    id: 'guide/core_concepts/smart-contract-anatomy'
                },
                {
                    type: 'doc',
                    label: 'Calling a Smart Contracts',
                    id: 'guide/core_concepts/invocation',
                },
                {
                    type: 'doc',
                    label: 'Sandbox Interface',
                    id: 'guide/core_concepts/sandbox'
                },
                {
                    type: 'category',
                    label: 'Core Contracts',
                    items: [
                        {
                            type: 'doc',
                            label: 'Overview',
                            id: 'guide/core_concepts/core_contracts/overview',
                        },
                        {
                            type: 'doc',
                            label: 'root',
                            id: 'guide/core_concepts/core_contracts/root',
                        },
                        {
                            type: 'doc',
                            label: 'accounts',
                            id: 'guide/core_concepts/core_contracts/accounts',
                        },
                        {
                            type: 'doc',
                            label: 'blob',
                            id: 'guide/core_concepts/core_contracts/blob',
                        },
                        {
                            type: 'doc',
                            label: 'blocklog',
                            id: 'guide/core_concepts/core_contracts/blocklog',
                        },
                        {
                            type: 'doc',
                            label: 'governance',
                            id: 'guide/core_concepts/core_contracts/governance',
                        },
                        {
                            type: 'doc',
                            label: 'errors',
                            id: 'guide/core_concepts/core_contracts/errors',
                        },
                        {
                            type: 'doc',
                            label: 'evm',
                            id: 'guide/core_concepts/core_contracts/evm',
                        },
                    ],
                },
                {
                    type: 'category',
                    label: 'Accounts',
                    items: [
                        {
                            type: 'doc',
                            label: 'How Accounts Work',
                            id: 'guide/core_concepts/accounts/how-accounts-work',
                        },
                        {
                            type: 'doc',
                            label: 'How To Deposit To a Chain',
                            id: 'guide/core_concepts/accounts/how-to-deposit-to-a-chain',
                        },
                        {
                            type: 'doc',
                            label: 'How To Withdraw From a Chain',
                            id: 'guide/core_concepts/accounts/how-to-withdraw-from-a-chain',
                        },
                        {
                            type: 'doc',
                            label: 'View Account Balances',
                            id: 'guide/core_concepts/accounts/view-account-balances',
                        },
                        {
                            type: 'doc',
                            label: 'The Common Account',
                            id: 'guide/core_concepts/accounts/the-common-account',
                        },
                    ]
                },
                {
                    type: 'category',
                    label: 'Testing Smart Contracts with Solo',
                    items: [
                        {
                            type: 'doc',
                            label: 'What is Solo?',
                            id: 'guide/solo/what-is-solo',
                        },
                        {
                            type: 'doc',
                            label: 'First Example',
                            id: 'guide/solo/first-example',
                        },
                        {
                            type: 'doc',
                            label: 'The L1 Ledger',
                            id: 'guide/solo/the-l1-ledger',
                        },
                        {
                            type: 'doc',
                            label: 'Deploying a Smart Contract',
                            id: 'guide/solo/deploying-sc',
                        },
                        {
                            type: 'doc',
                            label: 'Invoking a Smart Contract',
                            id: 'guide/solo/invoking-sc',
                        },
                        {
                            type: 'doc',
                            label: 'Calling a View',
                            id: 'guide/solo/view-sc',
                        },
                        {
                            type: 'doc',
                            label: 'Error Handling',
                            id: 'guide/solo/error-handling',
                        },
                        {
                            type: 'doc',
                            label: 'Accounts',
                            id: 'guide/solo/the-l2-ledger'
                        },
                    ]
                }
            ],
        },
        {
            type: 'category',
            label: 'ISC Chains and Nodes',
            items: [
                {
                    type: 'doc',
                    label: 'Installing Wasp',
                    id: 'guide/chains_and_nodes/installing-wasp',
                },
                {
                    type: 'doc',
                    label: 'Running a Node',
                    id: 'guide/chains_and_nodes/running-a-node',
                },
                {
                    type: 'doc',
                    label: 'Node Configuration',
                    id: 'guide/chains_and_nodes/node-config',
                },
                {
                    type: 'doc',
                    label: 'Using Docker (Standalone)',
                    id: 'guide/chains_and_nodes/docker_standalone',
                },
                {
                    type: 'doc',
                    label: 'Configuring wasp-cli',
                    id: 'guide/chains_and_nodes/wasp-cli',
                },
                {
                    type: 'doc',
                    label: 'Setting Up a Chain',
                    id: 'guide/chains_and_nodes/setting-up-a-chain',
                },
                {
                    type: 'doc',
                    label: 'Chain Management',
                    id: 'guide/chains_and_nodes/chain-management',
                },
                {
                    type: 'doc',
                    label: 'Testnet',
                    id: 'guide/chains_and_nodes/testnet',
                },
            ]
        },
        {
            type: 'category',
            label: 'Wasm VM',
            items: [
                {
                    type: 'category',
                    label: 'Introduction',
                    items: [
                        {
                            type: 'doc',
                            label: 'Wasm VM for ISC',
                            id: 'guide/wasm_vm/intro',
                        },
                        {
                            type: 'doc',
                            label: 'Data Access Proxies',
                            id: 'guide/wasm_vm/proxies',
                        },
                        {
                            type: 'doc',
                            label: 'WasmLib Data Types',
                            id: 'guide/wasm_vm/types',
                        },
                        {
                            type: 'doc',
                            label: 'Function Call Context',
                            id: 'guide/wasm_vm/context',
                        },
                    ]
                },
                {
                    type: 'category',
                    label: 'Schema tool',
                    items: [
                        {
                            type: 'doc',
                            label: 'Smart Contract Schema Tool',
                            id: 'guide/schema/schema',
                        },
                        {
                            type: 'doc',
                            label: 'Using the Schema Tool',
                            id: 'guide/schema/usage',
                        },
                        {
                            type: 'doc',
                            label: 'Structured Data Types',
                            id: 'guide/schema/structs',
                        },
                        {
                            type: 'doc',
                            label: 'Type Definitions',
                            id: 'guide/schema/typedefs',
                        },
                        {
                            type: 'doc',
                            label: 'Smart Contract State',
                            id: 'guide/schema/state',
                        },
                        {
                            type: 'doc',
                            label: 'Triggering Events',
                            id: 'guide/schema/events',
                        },
                        {
                            type: 'doc',
                            label: 'Function Definitions',
                            id: 'guide/schema/funcs',
                        },
                        {
                            type: 'doc',
                            label: 'Limiting Access',
                            id: 'guide/schema/access',
                        },
                        {
                            type: 'doc',
                            label: 'Function Parameters',
                            id: 'guide/schema/params',
                        },
                        {
                            type: 'doc',
                            label: 'Function Results',
                            id: 'guide/schema/results',
                        },
                        {
                            type: 'doc',
                            label: 'Thunk Functions',
                            id: 'guide/schema/thunks',
                        },
                        {
                            type: 'doc',
                            label: 'View-Only Functions',
                            id: 'guide/schema/views',
                        },
                        {
                            type: 'doc',
                            label: 'Smart Contract Initialization',
                            id: 'guide/schema/init',
                        },
                        {
                            type: 'doc',
                            label: 'Token Transfers',
                            id: 'guide/schema/transfers',
                        },
                        {
                            type: 'doc',
                            label: 'Function Descriptors',
                            id: 'guide/schema/funcdesc',
                        },
                        {
                            type: 'doc',
                            label: 'Calling Functions',
                            id: 'guide/schema/call',
                        },
                        {
                            type: 'doc',
                            label: 'Posting Asynchronous Requests',
                            id: 'guide/schema/post',
                        },
                        {
                            type: 'doc',
                            label: 'Testing Smart Contracts',
                            id: 'guide/schema/test',
                        },
                        {
                            type: 'doc',
                            label: 'Example Tests',
                            id: 'guide/schema/examples',
                        },
                        {
                            type: 'doc',
                            label: 'Colored Tokens and Time Locks',
                            id: 'guide/schema/timelock',
                        },
                    ]
                },
            ]
        },
        {
            type: 'category',
            label: 'EVM',
            items: [
                {
                    type: 'doc',
                    label: 'Introduction',
                    id: 'guide/evm/introduction',
                },
                {
                    type: 'doc',
                    label: 'Limitations',
                    id: 'guide/evm/compatibility',
                },
                {
                    type: 'doc',
                    label: 'How to use',
                    id: 'guide/evm/using',
                },
                {
                    type: 'doc',
                    label: 'The Magic Contract',
                    id: 'guide/evm/magic',
                },
                {
                    type: 'doc',
                    label: 'Tooling',
                    id: 'guide/evm/tooling',
                },
                {
                    type: 'category',
                    label: 'Examples',
                    items: [
                        {
                            type: 'doc',
                            label: 'Example Contract',
                            id: 'guide/evm/examples/introduction',
                        },
                        {
                            type: 'doc',
                            label: 'ERC20',
                            id: 'guide/evm/examples/ERC20',
                        },
                        {
                            type: 'doc',
                            label: 'ERC721',
                            id: 'guide/evm/examples/ERC721',
                        },
                    ]
                },
            ]
        },
        {
            type: 'category',
            label: 'Example projects',
            items: [
                {
                    type: 'doc',
                    label: 'Fair Roulette',
                    id: 'guide/example_projects/fair_roulette',
                },
            ]
        },
        {
            type: 'category',
            label: 'Development tools',
            items: [
                {
                    type: 'doc',
                    label: 'Preconfigured Development Docker setup',
                    id: 'guide/development_tools/docker_preconfigured',
                },
            ]
        },
        {
            type: 'doc',
            label: 'Contribute',
            id: 'contribute',
        },
        {
            type: 'doc',
            label: 'Metrics',
            id: 'metrics',
        }
    ],
};
