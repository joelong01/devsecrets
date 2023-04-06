#!/bin/bash

# Login to Azure using the Azure CLI
if ! az account show &>/dev/null; then
    # User is not logged in, so login to Azure using the Azure CLI
    az login &>/dev/null
fi

# Get the list of Azure subscriptions and their tenant IDs
subscriptions=$(az account list --query '[].{Name:name, SubscriptionId:id, TenantId:tenantId}' --output table)

# Print the list of subscriptions and their tenant IDs
echo "Here are your Azure subscriptions and their tenant IDs:"
echo "$subscriptions"
echo ""
echo -n "Please enter the Subscription ID you want to use: "
# Prompt the user to enter a subscription ID
read -r subscriptionId

# Return the subscription ID that the user picked - put a blank line so that the cli picks only the subscription
echo ""
echo "$subscriptionId"
