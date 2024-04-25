
/// Module: isc
module isc::anchor {
    use sui::{
        balance::{Self, Balance},
        sui::SUI, bag::{Self, Bag},
        coin::{Self, TreasuryCap, Coin},
        transfer::{Self, transfer, Receiving, receive},
        };
    use std::string::String;
    use std::type_name;

    /// privileged function was called without authorization
    const EWrongCaller: u64 = 1;

    /// A state controller
    public struct StateControllerCap has key, store{
        id: UID,
        index: u32,
        anchor_id: ID
    }

    /// A governor
    public struct GovernorCap has key, store{
        id: UID,
        index: u32,
        anchor_id: ID
    }

    /*	SenderContract ContractIdentity `json:"senderContract"`
	// ID of the target smart contract
	TargetContract Hname `json:"targetContract"`
	// entry point code
	EntryPoint Hname `json:"entryPoint"`
	// request arguments
	Params dict.Dict `json:"params"`
	// Allowance intended to the target contract to take. Nil means zero allowance
	Allowance *Assets `json:"allowance"`
	// gas budget
	GasBudget uint64 `json:"gasBudget"`
    */

    public struct Command<phantom T> has key,store {
        id: UID,
        targetContract: u64,
        entryPoint: u64,
        params: vector<vector<u8>>,
        allowance: Option<Balance<T>>,
        gasBudget: u64,
    }

    public struct Request<phantom T> has key, store {
        id: UID,
        gas_plus_base_token: Balance<SUI>,
        command: Command<T>,
        sender: address,
        asset: Option<Balance<T>>,
    }

    public struct Anchor has key, store{
        id: UID,
        /// iota holdings of the ISC chain
        base_token: Balance<SUI>,
        /// native token holdings of the ISC chain
        native_tokens: Bag,
        /// native nfts holdings of the ISC chain
        nfts: Bag,
        /// state index
        state_index: u32,
        // state metadata
        state_metadata: vector<u8>,
        // table that holds all the treasury caps of tokens minted by this chain
        minted_token_treasuries: Bag,

        current_state_controller: u32,
        current_governor: u32,
        metadata: vector<u8>,
    }

    /// sets up a new chain
    public fun start_new_chain(ctx: &mut TxContext): (Anchor, StateControllerCap, GovernorCap) {
        let anchor = Anchor{
            id: object::new(ctx),
            base_token: balance::zero(),
            native_tokens: bag::new(ctx),
            nfts: bag::new(ctx),
            state_index: 0,
            state_metadata: vector[],
            minted_token_treasuries: bag::new(ctx),
            current_state_controller: 0,
            current_governor: 0,
            metadata: vector[],
        };

        let state_controller = StateControllerCap{
            id: object::new(ctx),
            index: 0,
            anchor_id: anchor.id.uid_to_inner(),
        };

        let governor = GovernorCap{
            id: object::new(ctx),
            index: 0,
            anchor_id: anchor.id.uid_to_inner(),
        };

        (anchor, state_controller, governor)
    }

    /// PTB 1
    /// command 1: publish -> TreasuryCap, CoinMetadata
    ///
    /// PTB 2
    /// command 2: call register_isc_token(anchor, treasury, ctx)

    public fun register_isc_token<T>(anchor: &mut Anchor, treasury: TreasuryCap<T>) {
        let token_type: std::ascii::String = type_name::get<T>().into_string();
        anchor.minted_token_treasuries.add<std::ascii::String, TreasuryCap<T>>(token_type, treasury);
    }

    /// PTB
    /// call mint_token -> Coin
    /// call transfer coin somewhere

    public fun mint_token<T>(anchor: &mut Anchor, amount: u64, ctx: &mut TxContext): Coin<T> {
        let token_type: std::ascii::String = type_name::get<T>().into_string();
        let treasury = anchor.minted_token_treasuries.borrow_mut<std::ascii::String, TreasuryCap<T>>(token_type);

       treasury.mint(amount, ctx)
    }

    /**
           targetContract: u64,
        entryPoint: u64,
        params: vector<vector<u8>>,
        allowance: Option<Balance<T>>,
        gasBudget: u64,*/
    public fun create_command<T>(targetContract: u64, entryPoint: u64, params: vector<vector<u8>>, allowance: Option<Balance<T>>, gasBudget: u64, ctx: &mut TxContext): (Command<T>) {
        
        let command = Command<T>{
            id: object::new(ctx),
            targetContract: targetContract,
            entryPoint: entryPoint,
            params: params,
            allowance: allowance,
            gasBudget: gasBudget, 
        };

        command
    }

    /// clients call this to send a request to the anchor
    public fun send_request<T>(anchor: &Anchor, command: Command<T>, asset: Option<Balance<T>>, base_token: Balance<SUI>, ctx: &mut TxContext) {
        // we could implement spcific checks here for the request
        let request = Request<T>{
            id: object::new(ctx),
            gas_plus_base_token: base_token,
            command: command,
            sender: ctx.sender(),
            asset: asset,
        };

        // send the request object to the chain's id
        transfer(request, anchor.id.to_address());
    }

    public fun destroy_state_controller_cap(cap: StateControllerCap){
        let StateControllerCap {
            id: id,
            index: _,
            anchor_id: _,
        } = cap;
        object::delete(id);
    } 
    fun check_state_controller(anchor: &Anchor, caller: &StateControllerCap) {
        assert!(caller.anchor_id == anchor.id.uid_to_inner(), EWrongCaller);
        assert!(caller.index == anchor.current_state_controller, EWrongCaller);
    }

/*
    // function that consumes the request, has to be called by the state controller
    public fun consume<T>(caller : &StateControllerCap ,anchor: &mut Anchor, req: Receiving<Request<T>>, ctx: &mut TxContext) {
        check_state_controller(anchor, caller);

        let received_request = receive<Request<T>>(&mut anchor.id, req);

        // put the 


    }

*/
}

