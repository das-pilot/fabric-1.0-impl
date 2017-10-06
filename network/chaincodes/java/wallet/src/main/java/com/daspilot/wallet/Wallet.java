package com.daspilot.wallet;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.hyperledger.fabric.shim.ChaincodeBase;
import org.hyperledger.fabric.shim.ChaincodeStub;

/**
 * Wallet implementation
 * Created by Gleb Popov on 06-Oct-17.
 */
public class Wallet extends ChaincodeBase {

    private static Log log = LogFactory.getLog(Wallet.class);

    public static void main(String[] args) {
        log.info("Starting Wallet ChainCode");
        Wallet wallet = new Wallet();
        wallet.start(args);
    }

    @Override
    public Response init(ChaincodeStub chaincodeStub) {
        return newSuccessResponse("Ok");
    }

    @Override
    public Response invoke(ChaincodeStub chaincodeStub) {
        return newSuccessResponse("Ok");
    }
}
