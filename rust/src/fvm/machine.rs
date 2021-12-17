use std::collections::HashMap;
use std::convert::TryFrom;
use std::mem;
use std::ptr;
use std::sync::{atomic::AtomicU64, Mutex};

use anyhow::Error;
use blockstore::cgo::CgoBlockstore;
use blockstore::{Block, Blockstore, MemoryBlockstore};
use cid::Cid;
use drop_struct_macro_derive::DropStructMacro;
use ffi_toolkit::{
    c_str_to_pbuf, catch_panic_response, raw_ptr, rust_str_to_c_str, FCPResponseStatus,
};
use fvm::externs::cgo::CgoExterns;
use fvm::externs::Externs;
use fvm::machine::{ApplyKind, ApplyRet, Machine};
use fvm::message::Message;
use fvm::Config;
use fvm_shared::address::Address;
use fvm_shared::clock::ChainEpoch;
use fvm_shared::econ::TokenAmount;
use fvm_shared::encoding::RawBytes;
use fvm_shared::version::NetworkVersion;
use fvm_shared::MethodNum;
use log::{error, info};
use num_traits::FromPrimitive;
use once_cell::sync::Lazy;

use super::types::*;
use crate::util::api::init_log;

type CgoMachine = Machine<CgoBlockstore, CgoExterns>;

static FVM_MAP: Lazy<Mutex<HashMap<u64, CgoMachine>>> =
    Lazy::new(|| Mutex::new(HashMap::with_capacity(1)));

const NEXT_ID: AtomicU64 = AtomicU64::new(0);

fn add_fvm_machine(machine: CgoMachine) -> u64 {
    let next_id = NEXT_ID.fetch_add(1, std::sync::atomic::Ordering::SeqCst);
    let mut machines = FVM_MAP.lock().unwrap();
    machines.insert(next_id, machine);
    next_id
}

fn get_default_config() -> fvm::Config {
    Config {
        initial_pages: 1024, //FIXME
        max_pages: 32768,    // FIXME
        engine: wasmtime::Config::new(),
    }
}

/// Note: the incoming args as u64 and odd conversions to i32/i64
/// for some types is due to the generated bindings not liking the
/// 32bit types as incoming args
///
#[no_mangle]
#[cfg(not(target_os = "windows"))]
pub unsafe extern "C" fn fil_create_fvm_machine(
    fvm_version: fil_FvmRegisteredVersion,
    chain_epoch: u64,
    token_amount: u64,
    network_version: u64,
    state_root_ptr: *const u8,
    state_root_len: libc::size_t,
    blockstore_id: u64,
    externs_id: u64,
) -> *mut fil_CreateFvmMachineResponse {
    catch_panic_response(|| {
        init_log();

        info!("fil_create_fvm_machine: start");

        let mut response = fil_CreateFvmMachineResponse::default();

        let config = get_default_config();
        let chain_epoch = chain_epoch as ChainEpoch;
        let token_amount = TokenAmount::from_u64(token_amount);
        let token_amount = if token_amount.is_some() {
            token_amount.unwrap()
        } else {
            response.status_code = FCPResponseStatus::FCPUnclassifiedError;
            response.error_msg = rust_str_to_c_str(format!("token amount conversion failure"));
            return raw_ptr(response);
        };
        let network_version = match NetworkVersion::try_from(network_version as u32) {
            Ok(x) => x,
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
                return raw_ptr(response);
            }
        };
        let state_root_bytes: Vec<u8> =
            std::slice::from_raw_parts(state_root_ptr, state_root_len).to_vec();
        let state_root = match Cid::try_from(state_root_bytes) {
            Ok(x) => x,
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
                return raw_ptr(response);
            }
        };

        //let blockstore = MemoryBlockstore::new();
        let blockstore = CgoBlockstore::new(blockstore_id as i32);
        let externs = CgoExterns::new(externs_id as i32);
        let machine = fvm::machine::Machine::new(
            config,
            chain_epoch,
            token_amount,
            network_version,
            state_root,
            blockstore,
            externs,
        );
        match machine {
            Ok(machine) => {
                response.status_code = FCPResponseStatus::FCPNoError;
                response.machine_id = add_fvm_machine(machine);
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
                return raw_ptr(response);
            }
        }

        info!("fil_create_fvm_machine: finish");

        raw_ptr(response)
    })
}

#[no_mangle]
pub unsafe extern "C" fn fil_drop_fvm_machine(machine_id: u64) -> *mut fil_DropFvmMachineResponse {
    catch_panic_response(|| {
        init_log();

        info!("fil_drop_fvm_machine: start");

        let mut response = fil_DropFvmMachineResponse::default();

        let mut machines = FVM_MAP.lock().unwrap();
        let machine = machines.remove(&machine_id);
        match machine {
            Some(machine) => {
                response.status_code = FCPResponseStatus::FCPNoError;
            }
            None => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("invalid machine id"));
            }
        }

        info!("fil_drop_fvm_machine: end");

        raw_ptr(response)
    })
}

#[no_mangle]
pub unsafe extern "C" fn fil_fvm_machine_execute_message(
    machine_id: u64,
    message: fil_Message,
    apply_kind: u64, /* 0: Explicit, _: Implicit */
) -> *mut fil_FvmMachineExecuteResponse {
    catch_panic_response(|| {
        init_log();

        info!("fil_fvm_machine_execute_message: start");

        let mut response = fil_FvmMachineExecuteResponse::default();

        let apply_kind = if apply_kind == 0 {
            ApplyKind::Explicit
        } else {
            ApplyKind::Implicit
        };

        let message = match convert_fil_message_to_message(message) {
            Ok(x) => x,
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
                return raw_ptr(response);
            }
        };

        let mut machines = FVM_MAP.lock().unwrap();
        let mut machine = machines.get_mut(&machine_id);
        match machine {
            Some(machine) => {
                let apply_ret = match machine.execute_message(message, apply_kind) {
                    Ok(x) => x,
                    Err(err) => {
                        response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                        response.error_msg = rust_str_to_c_str(format!("{:?}", err));
                        return raw_ptr(response);
                    }
                };

                response.status_code = FCPResponseStatus::FCPNoError;
                // FIXME: Return relevant fields of ApplyRet
            }
            None => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("invalid machine id"));
            }
        }

        info!("fil_fvm_machine_execute_message: end");

        raw_ptr(response)
    })
}

#[no_mangle]
pub unsafe extern "C" fn fil_fvm_machine_finish_message(
    machine_id: u64,
    // TODO: actual message
) {
    // catch_panic_response(|| {
    init_log();

    info!("fil_fvm_machine_flush_message: start");

    let machines = FVM_MAP.lock().unwrap();
    let machine = machines.get(&machine_id);
    match machine {
        Some(machine) => {
            todo!("execute message")
        }
        None => {
            todo!("invalid machine id")
        }
    }

    info!("fil_fvm_machine_flush_message: end");
    // })
}

#[no_mangle]
pub unsafe extern "C" fn fil_destroy_create_fvm_machine_response(
    ptr: *mut fil_CreateFvmMachineResponse,
) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn fil_destroy_drop_fvm_machine_response(
    ptr: *mut fil_DropFvmMachineResponse,
) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn fil_destroy_fvm_machine_execute_response(
    ptr: *mut fil_FvmMachineExecuteResponse,
) {
    let _ = Box::from_raw(ptr);
}