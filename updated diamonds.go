package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"crypto/x509"
	"encoding/pem"
	"net/url"
	"regexp"
)

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4, 5
const   MINER           =  1
const   SHOPKEEPER      =  2
const   DEALER          =  3
const   BUYER           =  4
const   TRADER          =  5
const   CUTTER          =  6
const   JEWELLERY MAKER	=  7
const   CUSTOMER        =  8


//==============================================================================================================================
//	 Status types - Asset lifecycle is broken down into 5 statuses, this is part of the business logic to determine what can 
//					be done to the vehicle at points in it's lifecycle
//==============================================================================================================================
const   STATE_MINING  	        =  0
const   STATE_DISTRIBUTING	    =  1
const   STATE_INTER_DEALING     =  2
const   STATE_BUYING 	          =  3
const   STATE_TRADING           =  4      
const   STATE_CUTTING           =  5
const   STATE_JEWEL_MAKING      =  6
const   STATE_PURCHASING        =  7
         
  
//==============================================================================================================================
//	 Structure Definitions 
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}

//==============================================================================================================================
//      DIAMOND         - Defines the attributes of a diamond. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Diamond struct {
	Unique Id       int      `json:"Id"`
	Colour          string   `json:"colour"`
	Carat           int      `json:"carat"`
	Cut             string   `json:"cut"`					
	Clarity         string   `json:"clarity"`
	Location        string   `json:"location"`
	Date            int      `json:"date"`
	Stamp           time.Time`json:"stamp"`
	Polish          string   `json:"polish"`
	Symmetry        string   `json:"symmetry"`
        Jewellery Type  string   'json:"jewellery type"'
        Status          int      'json:"status"'
}


//==============================================================================================================================
//	Cutter                - Defines the structure that holds all the Unique Id's for diamonds that have been created.
//				Used as an index when querying all diamonds.
//==============================================================================================================================

type Cutter struct {
	Unique id []int `json:"Unique Id"`
}

//==============================================================================================================================
//	User_and_eCert - Struct for storing the JSON of a user and their ecert
//==============================================================================================================================

type User_and_eCert struct {
	Identity string `json:"identity"`
	eCert string `json:"ecert"`
}		

//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode																	
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	
	//Args
	//				0
	//			peer_address
	
	
	var Unique IDs Cutter
	
	bytes, err := json.Marshal(Unique IDs)
												if err != nil { return nil, errors.New("Error creating Cutter record") }
																
	err = stub.PutState("Unique IDs", bytes)
	
	for i:=0; i < len(args); i=i+2 {
		
		t.add_ecert(stub, args[i], args[i+1])													
	}

	return nil, nil
}

//==============================================================================================================================
//	 General Functions
//==============================================================================================================================
//	 get_ecert - Takes the name passed and calls out to the REST API for HyperLedger to retrieve the ecert
//				 for that user. Returns the ecert as retrived including html encoding.
//==============================================================================================================================
func (t *SimpleChaincode) get_ecert(stub *shim.ChaincodeStub, name string) ([]byte, error) {
	
	ecert, err := stub.GetState(name)

	if err != nil { return nil, errors.New("Couldn't retrieve ecert for user " + name) }
	
	return ecert, nil
}

//==============================================================================================================================
//	 add_ecert - Adds a new ecert and user pair to the table of ecerts
//==============================================================================================================================

func (t *SimpleChaincode) add_ecert(stub *shim.ChaincodeStub, name string, ecert string) ([]byte, error) {
	
	
	err := stub.PutState(name, []byte(ecert))

	if err == nil {
		return nil, errors.New("Error storing eCert for user " + name + " identity: " + ecert)
	}
	
	return nil, nil

}

//==============================================================================================================================
//	 get_caller      - Retrieves the username of the user who invoked the chaincode.
//				  Returns the username as a string.
//==============================================================================================================================

func (t *SimpleChaincode) get_username(stub *shim.ChaincodeStub) (string, error) {

	bytes, err := stub.GetCallerCertificate();
															if err != nil { return "", errors.New("Couldn't retrieve caller certificate") }
	x509Cert, err := x509.ParseCertificate(bytes);				// Extract Certificate from result of GetCallerCertificate						
															if err != nil { return "", errors.New("Couldn't parse certificate")	}
															
	return x509Cert.Subject.CommonName, nil
}

//==============================================================================================================================
//	 check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
// 				  		certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) check_affiliation(stub *shim.ChaincodeStub, cert string) (int, error) {																																																					
	

	decodedCert, err := url.QueryUnescape(cert);    				// make % etc normal //
	
															if err != nil { return -1, errors.New("Could not decode certificate") }
	
	pem, _ := pem.Decode([]byte(decodedCert))           				// Make Plain text   //

	x509Cert, err := x509.ParseCertificate(pem.Bytes);				// Extract Certificate from argument //
														
													if err != nil { return -1, errors.New("Couldn't parse certificate")	}

	cn := x509Cert.Subject.CommonName
	
	res := strings.Split(cn,"\\")
	
	affiliation, _ := strconv.Atoi(res[2])
	
	return affiliation, nil
		
}

//==============================================================================================================================
//	 get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *SimpleChaincode) get_caller_data(stub *shim.ChaincodeStub) (string, int, error){	

	user, err := t.get_username(stub)
																		if err != nil { return "", -1, err }
																		
	ecert, err := t.get_ecert(stub, user);					
																if err != nil { return "", -1, err }

	affiliation, err := t.check_affiliation(stub,string(ecert));			
																		if err != nil { return "", -1, err }

	return user, affiliation, nil
}

//==============================================================================================================================
//	 retrieve_Unique ID           - Gets the state of the data at Unique ID in the ledger then converts it from the stored 
//					JSON into the Diamond struct for use in the contract. Returns the Diamond struct.
//					Returns empty d if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_Unique ID(stub *shim.ChaincodeStub, Unique ID int) (Diamond, error) {
	
	var d Diamond

	bytes, err := stub.GetState(Unique ID);					
				
															if err != nil {	fmt.Printf("RETRIEVE_V5C: Failed to invoke vehicle_code: %s", err); return v, errors.New("RETRIEVE_V5C: Error retrieving vehicle with v5cID = " + v5cID) }

	err = json.Unmarshal(bytes, &v);						

															if err != nil {	fmt.Printf("RETRIEVE_V5C: Corrupt vehicle record "+string(bytes)+": %s", err); return v, errors.New("RETRIEVE_V5C: Corrupt vehicle record"+string(bytes))	}
	
	return v, nil
}

//==============================================================================================================================
// save_changes - Writes to the ledger the Vehicle struct passed in a JSON format. Uses the shim file's 
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub *shim.ChaincodeStub, d Diamond) (bool, error) {
	 
	bytes, err := json.Marshal(v)
	
																if err != nil { fmt.Printf("SAVE_CHANGES: Error converting vehicle record: %s", err); return false, errors.New("Error converting vehicle record") }

	err = stub.PutState(d.Unique IDs, bytes)
	
																if err != nil { fmt.Printf("SAVE_CHANGES: Error storing vehicle record: %s", err); return false, errors.New("Error storing vehicle record") }
	
	return true, nil
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	
	caller, caller_affiliation, err := t.get_caller_data(stub)

	if err != nil { return nil, errors.New("Error retrieving caller information")}

	
	if function == "create_diamond" { return t.create_diamond(stub, caller, caller_affiliation, args[0])
	} else { 																				// If the function is not a create then there must be a car so we need to retrieve the car.
		
		argPos := 1
		
		if function == "scrap_diamond" {																// If its a scrap vehicle then only two arguments are passed (no update value) all others have three arguments and the v5cID is expected in the last argument
			argPos = 0
		}
		
		d, err := t.retrieve_Unique ID(stub, args[argPos])
		
																							if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); return nil, errors.New("Error retrieving v5c") }
																		
		if strings.Contains(function, "update") == false           && 
		   function 							!= "scrap_diamond"    { 									// If the function is not an update or a scrappage it must be a transfer so we need to get the ecert of the recipient.
			
				ecert, err := t.get_ecert(stub, args[0]);					
				
																		if err != nil { return nil, err }

				rec_affiliation, err := t.check_affiliation(stub,string(ecert));	
				
																		if err != nil { return nil, err }
				
				if 		   function == "miner_to_shop_keeper" { return t.miner_to_shop_keeper(stub, d, caller, caller_affiliation, args[0], rec_affiliation)
				} else if  function == "shop_keeper_to_dealer"   { return t.shop_keeper_to_dealer(stub, d, caller, caller_affiliation, args[0], rec_affiliation)
				} else if  function == "dealer_to_buyer" 	   { return t.dealer_to_buyer(stub, d, caller, caller_affiliation, args[0], rec_affiliation)
				} else if  function == "buyer_to_trader"  { return t.buyer_to_trader(stub, d, caller, caller_affiliation, args[0], rec_affiliation)
				} else if  function == "trader_to_cutter"  { return t.trader_to_cutter(stub, d, caller, caller_affiliation, args[0], rec_affiliation)
				} else if  function == "cutter_to_jewellery_maker" { return t.cutter_to_jewellery_maker(stub, d, caller, caller_affiliation, args[0], rec_affiliation)
				} else if  function == "cutter_to_customer" { return t.cutter_to_customer(stub, d, caller, caller_affiliation, args[0], rec_affiliation)
                                }
			
		} else if function == "update_colour"  	    { return t.update_colour(stub, d, caller, caller_affiliation, args[0])
		} else if function == "update_cut"          { return t.update_model(stub, d, caller, caller_affiliation, args[0])
		} else if function == "update_clarity"   { return t.update_clarity(stub, d, caller, caller_affiliation, args[0])
		} else if function == "update_carat" 			{ return t.update_carat(stub, d, caller, caller_affiliation, args[0])
		} else if function == "update_symmetry" 		{ return t.update_symmetry(stub, d, caller, caller_affiliation, args[0])
		} else if function == "update_polish" 		{ return t.polish(stub, d, caller, caller_affiliation) }
		
																						return nil, errors.New("Function of that name doesn't exist.")
			
	}
}
//=================================================================================================================================	
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================	
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
													
	caller, caller_affiliation, err := t.get_caller_data(stub)

																							if err != nil { fmt.Printf("QUERY: Error retrieving caller details", err); return nil, errors.New("QUERY: Error retrieving caller details: "+err.Error()) }
															
	if function == "get_diamond_details" { 
	
			if len(args) != 1 { fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed") }
	
	
			v, err := t.retrieve_Unique ID(stub, args[0])
																							if err != nil { fmt.Printf("QUERY: Error retrieving v5c: %s", err); return nil, errors.New("QUERY: Error retrieving v5c "+err.Error()) }
	
			return t.get_diamond_details(stub, v, caller, caller_affiliation)
			
	} else if function == "get_diamonds" {
			return t.get_diamonds(stub, caller, caller_affiliation)
	} else if function == "get_ecert" {
			return t.get_ecert(stub, args[0])
	}

	return nil, errors.New("Received unknown function invocation")

}

//=================================================================================================================================
//	 Create Function
//=================================================================================================================================									
//	 Create Diamond - Creates the initial JSON for the diamond and then saves it to the ledger.									
//=================================================================================================================================
func (t *SimpleChaincode) create_diamond(stub *shim.ChaincodeStub, caller string, caller_affiliation int, UniqueID string) ([]byte, error) {								

	var d Diamond																																										
	
	Unique_ID      := "\"uniqueID\":\""+uniqueID+"\", "							// Variables to define the JSON
	colour         := "\"colour\":\"UNDEFINED\", "
	carat          := "\"carat\":\"UNDEFINED\", "
	cut            := "\"cut\":\"UNDEFINED\", "
	clarity        := "\"clarity\":\"UNDEFINED\", "
	location       := "\"location\":\""UNDEFINED"\", "
	date           := "\"date\":\"UNDEFINED\", "
	stamp          := "\"stamp\":\"UNDEFINED\", "
	polish         := "\"polish\":\"UNDEFINED\", "
	symmetry       := "\"Symmetry\":"UNDEFINED\", ""
        jewellery_type :="\"jewellery_type\":\"UNDEFINED\", "
        status         :="\"status\":0", "
	
	diamond_json := "{"+unique_ID+colour+carat+cut+clarity+location+date+stamp+polish+symmetry+jewellery_type+status"}" 	// Concatenates the variables to create the total JSON object
	
	matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(uniqueID))  				// matched = true if the v5cID passed fits format of two letters followed by seven digits
	
												if err != nil { fmt.Printf("CREATE_DIAMOND: Invalid uniqueID: %s", err); return nil, errors.New("Invalid uniqueID") }
	
	if 				unique_ID  == "" 	 || 
					matched == false    {
																		fmt.Printf("CREATE_DIAMOND: Invalid UNIQUEID provided");
																		return nil, errors.New("Invalid uniqueID provided")
	}

	err = json.Unmarshal([]byte(diamond_json), &d)							// Convert the JSON defined above into a diamond object for go
	
																		if err != nil { return nil, errors.New("Invalid JSON object") }

	record, err := stub.GetState(d.uniqueID) 								// If not an error then a record exists so cant create a new car with this uniqueID as it must be unique
	
																		if record != nil { return nil, errors.New("Diamond already exists") }
	
	if 	caller_affiliation != MINER {							// Only the regulator can create a new unique

																		return nil, errors.New("Permission Denied")
	}
	
	_, err  = t.save_changes(stub, d)									
			
																		if err != nil { fmt.Printf("CREATE_DIAMOND: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	bytes, err := stub.GetState("uniqueIDs")

																		if err != nil { return nil, errors.New("Unable to get uniqueIDs") }
																		
	var uniqueIDs cutter
	
	err = json.Unmarshal(bytes, &uniqueIDs)
	
																		if err != nil {	return nil, errors.New("Corrupt cutter record") }
															
	uniqueIDs.uniques = append(uniqueIDs.uniques, uniqueID)
	
	
	bytes, err = json.Marshal(uniqueIDs)
	
															if err != nil { fmt.Print("Error creating cutter record") }

	err = stub.PutState("uniqueIDs", bytes)

															if err != nil { return nil, errors.New("Unable to put the state") }
	
	return nil, nil

}

//=================================================================================================================================
//	 Transfer Functions
//=================================================================================================================================
//	 miner_to_shop_keeper
//=================================================================================================================================
func (t *SimpleChaincode) miner_to_shop_keeper(stub *shim.ChaincodeStub, d diamond, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	
if 		        d.Colour 	 == "UNDEFINED" || 					
			d.Carat          == "UNDEFINED" || 
			d.Stamp 	 == "TIMESTAMP" || 
			d.Date           == "UNDEFINED" || 
			d.Location       == "UNDEFINED"	||
 }
					



	if     	d.Status				== STATE_MINING	&&
			d.Owner					== caller			&&
			caller_affiliation		== MINER		&&
			recipient_affiliation	== SHOP_KEEPER		&&
			v.Scrapped				== false			{		// If the roles and users are ok 
	
					d.Owner  = recipient_name		// then make the owner the new owner
					d.Status = STATE_DISTRIBUTING			// and mark it in the state of manufacture
	
	} else {									// Otherwise if there is an error
	
															fmt.Printf(" MINER_TO_SHOPKEEPER: Permission Denied");
															return nil, errors.New("Permission Denied")
	
	}
	
	_, err := t.save_changes(stub, d)						// Write new state

															if err != nil {	fmt.Printf("MINER_TO_SHOPKEEPER: Error saving changes: %s", err); return nil, errors.New("Error saving changes")	}
														
	return nil, nil									// We are Done
	
}

//=================================================================================================================================
//	 shop_keeper_to_dealer
//=================================================================================================================================
func (t *SimpleChaincode) shop_keeper_to_dealer(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	
	
	
	if 		d.Status				== STATE_DISTRIBUTING	&& 
			d.Owner					== caller				&& 
			caller_affiliation		== SHOP_KEEPER			&&
			recipient_affiliation	== DEALER		&& 
			
					d.Owner = recipient_name
					d.Status = STATE_DEALING
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, d)
	
															if err != nil { fmt.Printf("SHOP_KEEPER_TO_DEALER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 dealer_to_buyer
//=================================================================================================================================
func (t *SimpleChaincode) dealer_to_buyer(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	
	if 		d.Status				== STATE_BUYING	&&
			d.Owner					== caller					&&
			caller_affiliation		== DEALER			&& 
			recipient_affiliation	== BUYER			&&
			
					d.Owner = recipient_name
					
	} else {
		
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_changes(stub, d)
	
															if err != nil { fmt.Printf("DEALER_TO_BUYER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 buyer_to_trader
//=================================================================================================================================
func (t *SimpleChaincode) buyer_to_trader(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	
	if 		d.Status				== STATE_TRADING	&& 
			d.Owner					== caller					&& 
			caller_affiliation		== BUYER			&& 
			recipient_affiliation	== TRADER			&& 
								{
		
					d.Owner = recipient_name
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, d)
															if err != nil { fmt.Printf("BUYER_TO_TRADER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 trader_to_cutter
//=================================================================================================================================
func (t *SimpleChaincode) trader_to_cutter(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	
if 		        d.Unique ID 	 == "UNDEFINED" || 					
			
 }


	if		d.Status				== STATE_CUTTING	&&
			d.Owner  				== caller					&& 
			caller_affiliation		== TRADER			&& 
			recipient_affiliation	== CUTTER			&& 
			v.Scrapped				== false					{
		
				d.Owner = recipient_name
	
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, d)
															if err != nil { fmt.Printf("TRADER_TO_CUTTER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 cutter_to_jewellery_maker
//=================================================================================================================================
func (t *SimpleChaincode) cutter_to_jewellery_maker(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	
if 		        d.Cut 	    == "UNDEFINED" || 					
			d.Symmetry  == "UNDEFINED" || 
                        d.Polish    == "UNDEFINED" || 
                        d.Unique Id == "UNDEFINED" || 
 }






	if		d.Status				== STATE_JEWEL_MAKING	&&
			d.Owner					== caller					&& 
			caller_affiliation		== CUTTER			&& 
			recipient_affiliation	== JEWELLERY_MAKER			&&
							{
			
					d.Owner = recipient_name
					
	} else {
		
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_changes(stub, d)
	
															if err != nil { fmt.Printf("CUTTER_TO_JEWELLERY_MAKER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}
/=================================================================================================================================
//	 jewellery_maker_to_customer
//=================================================================================================================================
func (t *SimpleChaincode) jewellery_maker_to_customer (stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	
if 		        d.Jewellery Type	    == "UNDEFINED" || 					
			
 }





	if		d.Status				== STATE_PURCHASING	&&
			d.Owner					== caller					&& 
			caller_affiliation		== JEWELLERY_MAKER			&& 
			recipient_affiliation	== CUSTOMER			&&
							{
			
					d.Owner = recipient_name
					
	} else {
		
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_changes(stub, d)
	
															if err != nil { fmt.Printf("JEWELLERY_MAKER_TO_CUSTOMER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 Update Functions
//=================================================================================================================================
//	 update_colour
//=================================================================================================================================
func (t *SimpleChaincode) update_colour(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	new_colour, err := strconv.Atoi(string(new_value)) 		                // will return an error if the new vin contains non numerical chars
	
															if err != nil || len(string(new_value)) != 15 { return nil, errors.New("Invalid value passed for new VIN") }
	
	if 		d.Status			== STATE_MINING	&& 
			d.Owner				== caller				&&
			caller_affiliation	== MINER			&&
			d.Colour									&&			// Can't change the colour after its initial assignment
			
					d.colour = new_colour					// Update to the new value
	} else {
	
															return nil, errors.New("Permission denied")
		
	}
	
	_, err  = t.save_changes(stub, d)						// Save the changes in the blockchain
	
															if err != nil { fmt.Printf("UPDATE_colour: Error saving changes: %s", err); return nil, errors.New("Error saving changes") } 
	
	return nil, nil
	
}


//=================================================================================================================================
//	 update_cut
//=================================================================================================================================
func (t *SimpleChaincode) update_cut(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int, new_value string) ([]byte, error) {

	
	if		d.Owner				== caller			&& 
			caller_affiliation	!=    CUTTER	&&
			
					d.Cut = new_value
	
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, d)
	
															if err != nil { fmt.Printf("UPDATE_CUT: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 update_clarity
//=================================================================================================================================
func (t *SimpleChaincode) update_colour(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	if 		v.Owner				== caller				&&
			caller_affiliation	== CUTTER			&&/*((v.Owner				== caller			&&
			caller_affiliation	== CUTTER		||
			
					d.Clarity = new_value
	} else {
	
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, d)
	
															if err != nil { fmt.Printf("UPDATE_CLARITY: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 update_CARAT
//=================================================================================================================================
func (t *SimpleChaincode) update_CARAT(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	if 		d.Status			== STATE_MINING	&&
			d.Owner				== caller				&& 
			caller_affiliation	== MINER		&&
			
					v.Make = new_value
	} else {
	
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_changes(stub, d)
	
															if err != nil { fmt.Printf("UPDATE_CARAT: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 update_SYMMETRY
//=================================================================================================================================
func (t *SimpleChaincode) update_symmetry(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	if 		d.Status			== STATE_CUTTING	&&
			d.Owner				== caller				&& 
			caller_affiliation	== CUTTER		&&
			
					d.Symmetry = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, d)
	
															if err != nil { fmt.Printf("UPDATE_SYMMETRY: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 update_POLISH
//=================================================================================================================================
func (t *SimpleChaincode) update_POLISH(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int) ([]byte, error) {

	if		d.Status			== STATE_CUTTING	&& 
			d.Owner				== caller				&& 
			caller_affiliation	== CUTTER		&& 
			
					
	} else {
		return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, d)
	
															if err != nil { fmt.Printf("SCRAP_VEHICLE: Error saving changes: %s", err); return nil, errors.New("SCRAP_VEHICLError saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 Read Functions
//=================================================================================================================================
//	 get_diamond_details
//=================================================================================================================================
func (t *SimpleChaincode) get_diamond_details(stub *shim.ChaincodeStub, d Diamond, caller string, caller_affiliation int) ([]byte, error) {
	
	bytes, err := json.Marshal(d)
	
																if err != nil { return nil, errors.New("GET_UNIQUE_DETAILS: Invalid diamond object") }
																
	if 		d.Owner				== caller		||
			caller_affiliation	== MINER	{
			
					return bytes, nil		
	} else {
																return nil, errors.New("Permission Denied")	
	}

}

//=================================================================================================================================
//	 get__diamond_details
//=================================================================================================================================

func (t *SimpleChaincode) get_diamonds(stub *shim.ChaincodeStub, caller string, caller_affiliation int) ([]byte, error) {

	bytes, err := stub.GetState("uniqueIDs")
		
																			if err != nil { return nil, errors.New("Unable to get uniqueIDs") }
																	
	var uniqueIDs cutter
	
	err = json.Unmarshal(bytes, &uniqueIDs)						
	
																			if err != nil {	return nil, errors.New("Corrupt Cutter") }
	
	result := "["
	
	var temp []byte
	var d Diamond
	
	for _, unique := range uniqueIDs.uniques {
		
		v, err = t.retrieve_unique(stub, unique)
		
		if err != nil {return nil, errors.New("Failed to retrieve Unique")}
		
		temp, err = t.get_diamond_details(stub, d, caller, caller_affiliation)
		
		if err == nil {
			result += string(temp) + ","	
		}
	}
	
	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}
	
	return []byte(result), nil
}

//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))
	
															if err != nil { fmt.Printf("Error starting Chaincode: %s", err) }
}
